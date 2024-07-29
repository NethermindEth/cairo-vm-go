%builtins range_check bitwise

from starkware.cairo.common.cairo_builtins import HashBuiltin, BitwiseBuiltin
from starkware.cairo.common.registers import get_label_location
from starkware.cairo.common.invoke import invoke
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.registers import get_fp_and_pc
from starkware.cairo.common.math import assert_nn_le, unsigned_div_rem
from starkware.cairo.common.math_cmp import is_le_felt
from starkware.cairo.common.memcpy import memcpy
from starkware.cairo.common.memset import memset
from starkware.cairo.common.pow import pow

const ALL_ONES = 2 ** 251 - 1;
const BLOCK_SIZE = 7;
const SHA256_INPUT_CHUNK_SIZE_FELTS = 16;
const SHA256_INPUT_CHUNK_SIZE_BYTES = 64;
const SHA256_STATE_SIZE_FELTS = 8;
const SHA256_INSTANCE_SIZE = SHA256_INPUT_CHUNK_SIZE_FELTS + 2 * SHA256_STATE_SIZE_FELTS;
const SHIFTS = 1 + 2 ** 35 + 2 ** (35 * 2) + 2 ** (35 * 3) + 2 ** (35 * 4) + 2 ** (35 * 5) + 2 ** (
    35 * 6
);

func sha256{range_check_ptr, sha256_ptr: felt*}(data: felt*, n_bytes: felt) -> (output: felt*) {
    alloc_locals;

    // Set the initial input state to IV.
    assert sha256_ptr[16] = 0x6A09E667;
    assert sha256_ptr[17] = 0xBB67AE85;
    assert sha256_ptr[18] = 0x3C6EF372;
    assert sha256_ptr[19] = 0xA54FF53A;
    assert sha256_ptr[20] = 0x510E527F;
    assert sha256_ptr[21] = 0x9B05688C;
    assert sha256_ptr[22] = 0x1F83D9AB;
    assert sha256_ptr[23] = 0x5BE0CD19;

    sha256_inner(data=data, n_bytes=n_bytes, total_bytes=n_bytes);

    // Set `output` to the start of the final state.
    let output = sha256_ptr;
    // Set `sha256_ptr` to the end of the output state.
    let sha256_ptr = sha256_ptr + SHA256_STATE_SIZE_FELTS;
    return (output,);
}

func _sha256_chunk{range_check_ptr, sha256_start: felt*, state: felt*, output: felt*}() {
    %{
        from starkware.cairo.common.cairo_sha256.sha256_utils import (
            compute_message_schedule, sha2_compress_function)

        _sha256_input_chunk_size_felts = int(ids.SHA256_INPUT_CHUNK_SIZE_FELTS)
        assert 0 <= _sha256_input_chunk_size_felts < 100
        _sha256_state_size_felts = int(ids.SHA256_STATE_SIZE_FELTS)
        assert 0 <= _sha256_state_size_felts < 100
        w = compute_message_schedule(memory.get_range(
            ids.sha256_start, _sha256_input_chunk_size_felts))
        new_state = sha2_compress_function(memory.get_range(ids.state, _sha256_state_size_felts), w)
        segments.write_arg(ids.output, new_state)
    %}
    return ();
}

func sha256_inner{range_check_ptr, sha256_ptr: felt*}(
    data: felt*, n_bytes: felt, total_bytes: felt
) {
    alloc_locals;

    let message = sha256_ptr;
    let state = sha256_ptr + SHA256_INPUT_CHUNK_SIZE_FELTS;
    let output = state + SHA256_STATE_SIZE_FELTS;

    let zero_bytes = is_le_felt(n_bytes, 0);
    let zero_total_bytes = is_le_felt(total_bytes, 0);

    // If the previous message block was full we are still missing "1" at the end of the message
    let (_, r_div_by_64) = unsigned_div_rem(total_bytes, 64);
    let missing_bit_one = is_le_felt(r_div_by_64, 0);

    // This works for 0 total bytes too, because zero_chunk will be -1 and, therefore, not 0.
    let zero_chunk = zero_bytes - zero_total_bytes - missing_bit_one;

    let is_last_block = is_le_felt(n_bytes, 55);
    if (is_last_block == 1) {
        _sha256_input(data, n_bytes, SHA256_INPUT_CHUNK_SIZE_FELTS - 2, zero_chunk);
        // Append the original message length at the end of the message block as a 64-bit big-endian integer.
        assert sha256_ptr[0] = 0;
        assert sha256_ptr[1] = total_bytes * 8;
        let sha256_ptr = sha256_ptr + 2;
        _sha256_chunk{sha256_start=message, state=state, output=output}();
        let sha256_ptr = sha256_ptr + SHA256_STATE_SIZE_FELTS;

        return ();
    }

    let (q, r) = unsigned_div_rem(n_bytes, SHA256_INPUT_CHUNK_SIZE_BYTES);
    let is_remainder_block = is_le_felt(q, 0);
    if (is_remainder_block == 1) {
        _sha256_input(data, r, SHA256_INPUT_CHUNK_SIZE_FELTS, 0);
        _sha256_chunk{sha256_start=message, state=state, output=output}();

        let sha256_ptr = sha256_ptr + SHA256_STATE_SIZE_FELTS;
        memcpy(
            output + SHA256_STATE_SIZE_FELTS + SHA256_INPUT_CHUNK_SIZE_FELTS,
            output,
            SHA256_STATE_SIZE_FELTS,
        );
        let sha256_ptr = sha256_ptr + SHA256_STATE_SIZE_FELTS;

        return sha256_inner(data=data, n_bytes=n_bytes - r, total_bytes=total_bytes);
    } else {
        _sha256_input(data, SHA256_INPUT_CHUNK_SIZE_BYTES, SHA256_INPUT_CHUNK_SIZE_FELTS, 0);
        _sha256_chunk{sha256_start=message, state=state, output=output}();

        let sha256_ptr = sha256_ptr + SHA256_STATE_SIZE_FELTS;
        memcpy(
            output + SHA256_STATE_SIZE_FELTS + SHA256_INPUT_CHUNK_SIZE_FELTS,
            output,
            SHA256_STATE_SIZE_FELTS,
        );
        let sha256_ptr = sha256_ptr + SHA256_STATE_SIZE_FELTS;

        return sha256_inner(
            data=data + SHA256_INPUT_CHUNK_SIZE_FELTS,
            n_bytes=n_bytes - SHA256_INPUT_CHUNK_SIZE_BYTES,
            total_bytes=total_bytes,
        );
    }
}

func _sha256_input{range_check_ptr, sha256_ptr: felt*}(
    input: felt*, n_bytes: felt, n_words: felt, pad_chunk: felt
) {
    alloc_locals;

    local full_word;
    %{ ids.full_word = int(ids.n_bytes >= 4) %}

    if (full_word != 0) {
        assert sha256_ptr[0] = input[0];
        let sha256_ptr = sha256_ptr + 1;
        return _sha256_input(
            input=input + 1, n_bytes=n_bytes - 4, n_words=n_words - 1, pad_chunk=pad_chunk
        );
    }

    if (n_words == 0) {
        return ();
    }

    if (n_bytes == 0 and pad_chunk == 1) {
        // Add zeros between the encoded message and the length integer so that the message block is a multiple of 512.
        memset(dst=sha256_ptr, value=0, n=n_words);
        let sha256_ptr = sha256_ptr + n_words;
        return ();
    }

    if (n_bytes == 0) {
        // This is the last input word, so we should add a byte '0x80' at the end and fill the rest with zeros.
        assert sha256_ptr[0] = 0x80000000;
        // Add zeros between the encoded message and the length integer so that the message block is a multiple of 512.
        memset(dst=sha256_ptr + 1, value=0, n=n_words - 1);
        let sha256_ptr = sha256_ptr + n_words;
        return ();
    }

    assert_nn_le(n_bytes, 3);
    let (padding) = pow(256, 3 - n_bytes);
    local range_check_ptr = range_check_ptr;

    assert sha256_ptr[0] = input[0] + padding * 0x80;

    memset(dst=sha256_ptr + 1, value=0, n=n_words - 1);
    let sha256_ptr = sha256_ptr + n_words;
    return ();
}

func compute_message_schedule{bitwise_ptr: BitwiseBuiltin*}(message: felt*) {
    alloc_locals;

    // Defining the following constants as local variables saves some instructions.
    local shift_mask3 = SHIFTS * (2 ** 32 - 2 ** 3);
    local shift_mask7 = SHIFTS * (2 ** 32 - 2 ** 7);
    local shift_mask10 = SHIFTS * (2 ** 32 - 2 ** 10);
    local shift_mask17 = SHIFTS * (2 ** 32 - 2 ** 17);
    local shift_mask18 = SHIFTS * (2 ** 32 - 2 ** 18);
    local shift_mask19 = SHIFTS * (2 ** 32 - 2 ** 19);
    local mask32ones = SHIFTS * (2 ** 32 - 1);

    // Loop variables.
    tempvar bitwise_ptr = bitwise_ptr;
    tempvar message = message + 16;
    tempvar n = 64 - 16;

    loop:
    // Compute s0 = right_rot(w[i - 15], 7) ^ right_rot(w[i - 15], 18) ^ (w[i - 15] >> 3).
    tempvar w0 = message[-15];
    assert bitwise_ptr[0].x = w0;
    assert bitwise_ptr[0].y = shift_mask7;
    let w0_rot7 = (2 ** (32 - 7)) * w0 + (1 / 2 ** 7 - 2 ** (32 - 7)) * bitwise_ptr[0].x_and_y;
    assert bitwise_ptr[1].x = w0;
    assert bitwise_ptr[1].y = shift_mask18;
    let w0_rot18 = (2 ** (32 - 18)) * w0 + (1 / 2 ** 18 - 2 ** (32 - 18)) * bitwise_ptr[1].x_and_y;
    assert bitwise_ptr[2].x = w0;
    assert bitwise_ptr[2].y = shift_mask3;
    let w0_shift3 = (1 / 2 ** 3) * bitwise_ptr[2].x_and_y;
    assert bitwise_ptr[3].x = w0_rot7;
    assert bitwise_ptr[3].y = w0_rot18;
    assert bitwise_ptr[4].x = bitwise_ptr[3].x_xor_y;
    assert bitwise_ptr[4].y = w0_shift3;
    let s0 = bitwise_ptr[4].x_xor_y;
    let bitwise_ptr = bitwise_ptr + 5 * BitwiseBuiltin.SIZE;

    // Compute s1 = right_rot(w[i - 2], 17) ^ right_rot(w[i - 2], 19) ^ (w[i - 2] >> 10).
    tempvar w1 = message[-2];
    assert bitwise_ptr[0].x = w1;
    assert bitwise_ptr[0].y = shift_mask17;
    let w1_rot17 = (2 ** (32 - 17)) * w1 + (1 / 2 ** 17 - 2 ** (32 - 17)) * bitwise_ptr[0].x_and_y;
    assert bitwise_ptr[1].x = w1;
    assert bitwise_ptr[1].y = shift_mask19;
    let w1_rot19 = (2 ** (32 - 19)) * w1 + (1 / 2 ** 19 - 2 ** (32 - 19)) * bitwise_ptr[1].x_and_y;
    assert bitwise_ptr[2].x = w1;
    assert bitwise_ptr[2].y = shift_mask10;
    let w1_shift10 = (1 / 2 ** 10) * bitwise_ptr[2].x_and_y;
    assert bitwise_ptr[3].x = w1_rot17;
    assert bitwise_ptr[3].y = w1_rot19;
    assert bitwise_ptr[4].x = bitwise_ptr[3].x_xor_y;
    assert bitwise_ptr[4].y = w1_shift10;
    let s1 = bitwise_ptr[4].x_xor_y;
    let bitwise_ptr = bitwise_ptr + 5 * BitwiseBuiltin.SIZE;

    assert bitwise_ptr[0].x = message[-16] + s0 + message[-7] + s1;
    assert bitwise_ptr[0].y = mask32ones;
    assert message[0] = bitwise_ptr[0].x_and_y;
    let bitwise_ptr = bitwise_ptr + BitwiseBuiltin.SIZE;

    tempvar bitwise_ptr = bitwise_ptr;
    tempvar message = message + 1;
    tempvar n = n - 1;
    jmp loop if n != 0;

    return ();
}

func sha2_compress{bitwise_ptr: BitwiseBuiltin*}(
    state: felt*, message: felt*, round_constants: felt*
) -> (new_state: felt*) {
    alloc_locals;

    // Defining the following constants as local variables saves some instructions.
    local shift_mask2 = SHIFTS * (2 ** 32 - 2 ** 2);
    local shift_mask13 = SHIFTS * (2 ** 32 - 2 ** 13);
    local shift_mask22 = SHIFTS * (2 ** 32 - 2 ** 22);
    local shift_mask6 = SHIFTS * (2 ** 32 - 2 ** 6);
    local shift_mask11 = SHIFTS * (2 ** 32 - 2 ** 11);
    local shift_mask25 = SHIFTS * (2 ** 32 - 2 ** 25);
    local mask32ones = SHIFTS * (2 ** 32 - 1);

    tempvar a = state[0];
    tempvar b = state[1];
    tempvar c = state[2];
    tempvar d = state[3];
    tempvar e = state[4];
    tempvar f = state[5];
    tempvar g = state[6];
    tempvar h = state[7];
    tempvar round_constants = round_constants;
    tempvar message = message;
    tempvar bitwise_ptr = bitwise_ptr;
    tempvar n = 64;

    loop:
    // Compute s0 = right_rot(a, 2) ^ right_rot(a, 13) ^ right_rot(a, 22).
    assert bitwise_ptr[0].x = a;
    assert bitwise_ptr[0].y = shift_mask2;
    let a_rot2 = (2 ** (32 - 2)) * a + (1 / 2 ** 2 - 2 ** (32 - 2)) * bitwise_ptr[0].x_and_y;
    assert bitwise_ptr[1].x = a;
    assert bitwise_ptr[1].y = shift_mask13;
    let a_rot13 = (2 ** (32 - 13)) * a + (1 / 2 ** 13 - 2 ** (32 - 13)) * bitwise_ptr[1].x_and_y;
    assert bitwise_ptr[2].x = a;
    assert bitwise_ptr[2].y = shift_mask22;
    let a_rot22 = (2 ** (32 - 22)) * a + (1 / 2 ** 22 - 2 ** (32 - 22)) * bitwise_ptr[2].x_and_y;
    assert bitwise_ptr[3].x = a_rot2;
    assert bitwise_ptr[3].y = a_rot13;
    assert bitwise_ptr[4].x = bitwise_ptr[3].x_xor_y;
    assert bitwise_ptr[4].y = a_rot22;
    let s0 = bitwise_ptr[4].x_xor_y;
    let bitwise_ptr = bitwise_ptr + 5 * BitwiseBuiltin.SIZE;

    // Compute s1 = right_rot(e, 6) ^ right_rot(e, 11) ^ right_rot(e, 25).
    assert bitwise_ptr[0].x = e;
    assert bitwise_ptr[0].y = shift_mask6;
    let e_rot6 = (2 ** (32 - 6)) * e + (1 / 2 ** 6 - 2 ** (32 - 6)) * bitwise_ptr[0].x_and_y;
    assert bitwise_ptr[1].x = e;
    assert bitwise_ptr[1].y = shift_mask11;
    let e_rot11 = (2 ** (32 - 11)) * e + (1 / 2 ** 11 - 2 ** (32 - 11)) * bitwise_ptr[1].x_and_y;
    assert bitwise_ptr[2].x = e;
    assert bitwise_ptr[2].y = shift_mask25;
    let e_rot25 = (2 ** (32 - 25)) * e + (1 / 2 ** 25 - 2 ** (32 - 25)) * bitwise_ptr[2].x_and_y;
    assert bitwise_ptr[3].x = e_rot6;
    assert bitwise_ptr[3].y = e_rot11;
    assert bitwise_ptr[4].x = bitwise_ptr[3].x_xor_y;
    assert bitwise_ptr[4].y = e_rot25;
    let s1 = bitwise_ptr[4].x_xor_y;
    let bitwise_ptr = bitwise_ptr + 5 * BitwiseBuiltin.SIZE;

    // Compute ch = (e & f) ^ ((~e) & g).
    assert bitwise_ptr[0].x = e;
    assert bitwise_ptr[0].y = f;
    assert bitwise_ptr[1].x = ALL_ONES - e;
    assert bitwise_ptr[1].y = g;
    let ch = bitwise_ptr[0].x_and_y + bitwise_ptr[1].x_and_y;
    let bitwise_ptr = bitwise_ptr + 2 * BitwiseBuiltin.SIZE;

    // Compute maj = (a & b) ^ (a & c) ^ (b & c).
    assert bitwise_ptr[0].x = a;
    assert bitwise_ptr[0].y = b;
    assert bitwise_ptr[1].x = bitwise_ptr[0].x_xor_y;
    assert bitwise_ptr[1].y = c;
    let maj = (a + b + c - bitwise_ptr[1].x_xor_y) / 2;
    let bitwise_ptr = bitwise_ptr + 2 * BitwiseBuiltin.SIZE;

    tempvar temp1 = h + s1 + ch + round_constants[0] + message[0];
    tempvar temp2 = s0 + maj;

    assert bitwise_ptr[0].x = temp1 + temp2;
    assert bitwise_ptr[0].y = mask32ones;
    let new_a = bitwise_ptr[0].x_and_y;
    assert bitwise_ptr[1].x = d + temp1;
    assert bitwise_ptr[1].y = mask32ones;
    let new_e = bitwise_ptr[1].x_and_y;
    let bitwise_ptr = bitwise_ptr + 2 * BitwiseBuiltin.SIZE;

    tempvar new_a = new_a;
    tempvar new_b = a;
    tempvar new_c = b;
    tempvar new_d = c;
    tempvar new_e = new_e;
    tempvar new_f = e;
    tempvar new_g = f;
    tempvar new_h = g;
    tempvar round_constants = round_constants + 1;
    tempvar message = message + 1;
    tempvar bitwise_ptr = bitwise_ptr;
    tempvar n = n - 1;
    jmp loop if n != 0;

    // Add the compression result to the original state:
    let (res) = alloc();
    assert bitwise_ptr[0].x = state[0] + new_a;
    assert bitwise_ptr[0].y = mask32ones;
    assert res[0] = bitwise_ptr[0].x_and_y;
    assert bitwise_ptr[1].x = state[1] + new_b;
    assert bitwise_ptr[1].y = mask32ones;
    assert res[1] = bitwise_ptr[1].x_and_y;
    assert bitwise_ptr[2].x = state[2] + new_c;
    assert bitwise_ptr[2].y = mask32ones;
    assert res[2] = bitwise_ptr[2].x_and_y;
    assert bitwise_ptr[3].x = state[3] + new_d;
    assert bitwise_ptr[3].y = mask32ones;
    assert res[3] = bitwise_ptr[3].x_and_y;
    assert bitwise_ptr[4].x = state[4] + new_e;
    assert bitwise_ptr[4].y = mask32ones;
    assert res[4] = bitwise_ptr[4].x_and_y;
    assert bitwise_ptr[5].x = state[5] + new_f;
    assert bitwise_ptr[5].y = mask32ones;
    assert res[5] = bitwise_ptr[5].x_and_y;
    assert bitwise_ptr[6].x = state[6] + new_g;
    assert bitwise_ptr[6].y = mask32ones;
    assert res[6] = bitwise_ptr[6].x_and_y;
    assert bitwise_ptr[7].x = state[7] + new_h;
    assert bitwise_ptr[7].y = mask32ones;
    assert res[7] = bitwise_ptr[7].x_and_y;
    let bitwise_ptr = bitwise_ptr + 8 * BitwiseBuiltin.SIZE;

    return (res,);
}

// Returns the 64 round constants of SHA256.
func get_round_constants() -> (round_constants: felt*) {
    alloc_locals;
    let (__fp__, _) = get_fp_and_pc();
    local round_constants = 0x428A2F98 * SHIFTS;
    local a = 0x71374491 * SHIFTS;
    local a = 0xB5C0FBCF * SHIFTS;
    local a = 0xE9B5DBA5 * SHIFTS;
    local a = 0x3956C25B * SHIFTS;
    local a = 0x59F111F1 * SHIFTS;
    local a = 0x923F82A4 * SHIFTS;
    local a = 0xAB1C5ED5 * SHIFTS;
    local a = 0xD807AA98 * SHIFTS;
    local a = 0x12835B01 * SHIFTS;
    local a = 0x243185BE * SHIFTS;
    local a = 0x550C7DC3 * SHIFTS;
    local a = 0x72BE5D74 * SHIFTS;
    local a = 0x80DEB1FE * SHIFTS;
    local a = 0x9BDC06A7 * SHIFTS;
    local a = 0xC19BF174 * SHIFTS;
    local a = 0xE49B69C1 * SHIFTS;
    local a = 0xEFBE4786 * SHIFTS;
    local a = 0x0FC19DC6 * SHIFTS;
    local a = 0x240CA1CC * SHIFTS;
    local a = 0x2DE92C6F * SHIFTS;
    local a = 0x4A7484AA * SHIFTS;
    local a = 0x5CB0A9DC * SHIFTS;
    local a = 0x76F988DA * SHIFTS;
    local a = 0x983E5152 * SHIFTS;
    local a = 0xA831C66D * SHIFTS;
    local a = 0xB00327C8 * SHIFTS;
    local a = 0xBF597FC7 * SHIFTS;
    local a = 0xC6E00BF3 * SHIFTS;
    local a = 0xD5A79147 * SHIFTS;
    local a = 0x06CA6351 * SHIFTS;
    local a = 0x14292967 * SHIFTS;
    local a = 0x27B70A85 * SHIFTS;
    local a = 0x2E1B2138 * SHIFTS;
    local a = 0x4D2C6DFC * SHIFTS;
    local a = 0x53380D13 * SHIFTS;
    local a = 0x650A7354 * SHIFTS;
    local a = 0x766A0ABB * SHIFTS;
    local a = 0x81C2C92E * SHIFTS;
    local a = 0x92722C85 * SHIFTS;
    local a = 0xA2BFE8A1 * SHIFTS;
    local a = 0xA81A664B * SHIFTS;
    local a = 0xC24B8B70 * SHIFTS;
    local a = 0xC76C51A3 * SHIFTS;
    local a = 0xD192E819 * SHIFTS;
    local a = 0xD6990624 * SHIFTS;
    local a = 0xF40E3585 * SHIFTS;
    local a = 0x106AA070 * SHIFTS;
    local a = 0x19A4C116 * SHIFTS;
    local a = 0x1E376C08 * SHIFTS;
    local a = 0x2748774C * SHIFTS;
    local a = 0x34B0BCB5 * SHIFTS;
    local a = 0x391C0CB3 * SHIFTS;
    local a = 0x4ED8AA4A * SHIFTS;
    local a = 0x5B9CCA4F * SHIFTS;
    local a = 0x682E6FF3 * SHIFTS;
    local a = 0x748F82EE * SHIFTS;
    local a = 0x78A5636F * SHIFTS;
    local a = 0x84C87814 * SHIFTS;
    local a = 0x8CC70208 * SHIFTS;
    local a = 0x90BEFFFA * SHIFTS;
    local a = 0xA4506CEB * SHIFTS;
    local a = 0xBEF9A3F7 * SHIFTS;
    local a = 0xC67178F2 * SHIFTS;
    return (&round_constants,);
}

// Handles n blocks of BLOCK_SIZE SHA256 instances.
// Taken from: https://github.com/starkware-libs/cairo-examples/blob/0d88b41bffe3de112d98986b8b0afa795f9d67a0/sha256/sha256.cairo#L102
func _finalize_sha256_inner{range_check_ptr, bitwise_ptr: BitwiseBuiltin*}(
    sha256_ptr: felt*, n: felt, round_constants: felt*
) {
    if (n == 0) {
        return ();
    }

    alloc_locals;

    local MAX_VALUE = 2 ** 32 - 1;

    let sha256_start = sha256_ptr;

    let (local message_start: felt*) = alloc();
    let (local input_state_start: felt*) = alloc();

    // Handle message.

    tempvar message = message_start;
    tempvar sha256_ptr = sha256_ptr;
    tempvar range_check_ptr = range_check_ptr;
    tempvar m = SHA256_INPUT_CHUNK_SIZE_FELTS;

    message_loop:
    tempvar x0 = sha256_ptr[0 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 0] = x0;
    assert [range_check_ptr + 1] = MAX_VALUE - x0;
    tempvar x1 = sha256_ptr[1 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 2] = x1;
    assert [range_check_ptr + 3] = MAX_VALUE - x1;
    tempvar x2 = sha256_ptr[2 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 4] = x2;
    assert [range_check_ptr + 5] = MAX_VALUE - x2;
    tempvar x3 = sha256_ptr[3 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 6] = x3;
    assert [range_check_ptr + 7] = MAX_VALUE - x3;
    tempvar x4 = sha256_ptr[4 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 8] = x4;
    assert [range_check_ptr + 9] = MAX_VALUE - x4;
    tempvar x5 = sha256_ptr[5 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 10] = x5;
    assert [range_check_ptr + 11] = MAX_VALUE - x5;
    tempvar x6 = sha256_ptr[6 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 12] = x6;
    assert [range_check_ptr + 13] = MAX_VALUE - x6;
    assert message[0] = x0 + 2 ** 35 * x1 + 2 ** (35 * 2) * x2 + 2 ** (35 * 3) * x3 + 2 ** (
        35 * 4
    ) * x4 + 2 ** (35 * 5) * x5 + 2 ** (35 * 6) * x6;

    tempvar message = message + 1;
    tempvar sha256_ptr = sha256_ptr + 1;
    tempvar range_check_ptr = range_check_ptr + 14;
    tempvar m = m - 1;
    jmp message_loop if m != 0;

    // Handle input state.

    tempvar input_state = input_state_start;
    tempvar sha256_ptr = sha256_ptr;
    tempvar range_check_ptr = range_check_ptr;
    tempvar m = SHA256_STATE_SIZE_FELTS;

    input_state_loop:
    tempvar x0 = sha256_ptr[0 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 0] = x0;
    assert [range_check_ptr + 1] = MAX_VALUE - x0;
    tempvar x1 = sha256_ptr[1 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 2] = x1;
    assert [range_check_ptr + 3] = MAX_VALUE - x1;
    tempvar x2 = sha256_ptr[2 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 4] = x2;
    assert [range_check_ptr + 5] = MAX_VALUE - x2;
    tempvar x3 = sha256_ptr[3 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 6] = x3;
    assert [range_check_ptr + 7] = MAX_VALUE - x3;
    tempvar x4 = sha256_ptr[4 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 8] = x4;
    assert [range_check_ptr + 9] = MAX_VALUE - x4;
    tempvar x5 = sha256_ptr[5 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 10] = x5;
    assert [range_check_ptr + 11] = MAX_VALUE - x5;
    tempvar x6 = sha256_ptr[6 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 12] = x6;
    assert [range_check_ptr + 13] = MAX_VALUE - x6;
    assert input_state[0] = x0 + 2 ** 35 * x1 + 2 ** (35 * 2) * x2 + 2 ** (35 * 3) * x3 + 2 ** (
        35 * 4
    ) * x4 + 2 ** (35 * 5) * x5 + 2 ** (35 * 6) * x6;

    tempvar input_state = input_state + 1;
    tempvar sha256_ptr = sha256_ptr + 1;
    tempvar range_check_ptr = range_check_ptr + 14;
    tempvar m = m - 1;
    jmp input_state_loop if m != 0;

    // Run sha256 on the 7 instances.

    local sha256_ptr: felt* = sha256_ptr;
    local range_check_ptr = range_check_ptr;
    compute_message_schedule(message_start);
    let (outputs) = sha2_compress(input_state_start, message_start, round_constants);
    local bitwise_ptr: BitwiseBuiltin* = bitwise_ptr;

    // Handle outputs.

    tempvar outputs = outputs;
    tempvar sha256_ptr = sha256_ptr;
    tempvar range_check_ptr = range_check_ptr;
    tempvar m = SHA256_STATE_SIZE_FELTS;

    output_loop:
    tempvar x0 = sha256_ptr[0 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr] = x0;
    assert [range_check_ptr + 1] = MAX_VALUE - x0;
    tempvar x1 = sha256_ptr[1 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 2] = x1;
    assert [range_check_ptr + 3] = MAX_VALUE - x1;
    tempvar x2 = sha256_ptr[2 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 4] = x2;
    assert [range_check_ptr + 5] = MAX_VALUE - x2;
    tempvar x3 = sha256_ptr[3 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 6] = x3;
    assert [range_check_ptr + 7] = MAX_VALUE - x3;
    tempvar x4 = sha256_ptr[4 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 8] = x4;
    assert [range_check_ptr + 9] = MAX_VALUE - x4;
    tempvar x5 = sha256_ptr[5 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 10] = x5;
    assert [range_check_ptr + 11] = MAX_VALUE - x5;
    tempvar x6 = sha256_ptr[6 * SHA256_INSTANCE_SIZE];
    assert [range_check_ptr + 12] = x6;
    assert [range_check_ptr + 13] = MAX_VALUE - x6;

    assert outputs[0] = x0 + 2 ** 35 * x1 + 2 ** (35 * 2) * x2 + 2 ** (35 * 3) * x3 + 2 ** (
        35 * 4
    ) * x4 + 2 ** (35 * 5) * x5 + 2 ** (35 * 6) * x6;

    tempvar outputs = outputs + 1;
    tempvar sha256_ptr = sha256_ptr + 1;
    tempvar range_check_ptr = range_check_ptr + 14;
    tempvar m = m - 1;
    jmp output_loop if m != 0;

    return _finalize_sha256_inner(
        sha256_ptr=sha256_start + SHA256_INSTANCE_SIZE * BLOCK_SIZE,
        n=n - 1,
        round_constants=round_constants,
    );
}

// Verifies that the results of sha256() are valid.
// Taken from: https://github.com/starkware-libs/cairo-examples/blob/0d88b41bffe3de112d98986b8b0afa795f9d67a0/sha256/sha256.cairo#L246
func finalize_sha256{range_check_ptr, bitwise_ptr: BitwiseBuiltin*}(
    sha256_ptr_start: felt*, sha256_ptr_end: felt*
) {
    alloc_locals;

    let (__fp__, _) = get_fp_and_pc();

    let (round_constants) = get_round_constants();

    // We reuse the output state of the previous chunk as input to the next.
    tempvar n = (sha256_ptr_end - sha256_ptr_start) / SHA256_INSTANCE_SIZE;
    if (n == 0) {
        return ();
    }

    %{
        # Add dummy pairs of input and output.
        from starkware.cairo.common.cairo_sha256.sha256_utils import (
            IV, compute_message_schedule, sha2_compress_function)

        _block_size = int(ids.BLOCK_SIZE)
        assert 0 <= _block_size < 20
        _sha256_input_chunk_size_felts = int(ids.SHA256_INPUT_CHUNK_SIZE_FELTS)
        assert 0 <= _sha256_input_chunk_size_felts < 100

        message = [0] * _sha256_input_chunk_size_felts
        w = compute_message_schedule(message)
        output = sha2_compress_function(IV, w)
        padding = (message + IV + output) * (_block_size - 1)
        segments.write_arg(ids.sha256_ptr_end, padding)
    %}

    // Compute the amount of blocks (rounded up).
    let (local q, r) = unsigned_div_rem(n + BLOCK_SIZE - 1, BLOCK_SIZE);
    _finalize_sha256_inner(sha256_ptr_start, n=q, round_constants=round_constants);
    return ();
}

// Taken from https://github.com/cartridge-gg/cairo-sha256/blob/8d2ae515ab5cc9fc530c2dcf3ed1172bd181136e/tests/test_sha256.cairo
func test_sha256_hello_world{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;

    let (hello_world) = alloc();
    assert hello_world[0] = 'hell';
    assert hello_world[1] = 'o wo';
    assert hello_world[2] = 'rld\x00';

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(hello_world, 11);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    assert hash[0] = 3108841401;
    assert hash[1] = 2471312904;
    assert hash[2] = 2771276503;
    assert hash[3] = 3665669114;
    assert hash[4] = 3297046499;
    assert hash[5] = 2052292846;
    assert hash[6] = 2424895404;
    assert hash[7] = 3807366633;

    return ();
}

func test_sha256_multichunks{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;

    let (phrase) = alloc();
    // phrase="this is an example message which should take multiple chunks"
    // 01110100 01101000 01101001 01110011
    assert phrase[0] = 1952999795;
    // 00100000 01101001 01110011 00100000
    assert phrase[1] = 543781664;
    // 01100001 01101110 00100000 01100101
    assert phrase[2] = 1634607205;
    // 01111000 01100001 01101101 01110000
    assert phrase[3] = 2019650928;
    // 01101100 01100101 00100000 01101101
    assert phrase[4] = 1818566765;
    // 01100101 01110011 01110011 01100001
    assert phrase[5] = 1702064993;
    // 01100111 01100101 00100000 01110111
    assert phrase[6] = 1734680695;
    // 01101000 01101001 01100011 01101000
    assert phrase[7] = 1751737192;
    // 00100000 01110011 01101000 01101111
    assert phrase[8] = 544434287;
    // 01110101 01101100 01100100 00100000
    assert phrase[9] = 1970037792;
    // 01110100 01100001 01101011 01100101
    assert phrase[10] = 1952541541;
    // 00100000 01101101 01110101 01101100
    assert phrase[11] = 544044396;
    // 01110100 01101001 01110000 01101100
    assert phrase[12] = 1953067116;
    // 01100101 00100000 01100011 01101000
    assert phrase[13] = 1696621416;
    // 01110101 01101110 01101011 01110011
    assert phrase[14] = 1970170739;

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(phrase, 60);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    assert hash[0] = 3714276112;
    assert hash[1] = 759782134;
    assert hash[2] = 1331117438;
    assert hash[3] = 1287649296;
    assert hash[4] = 699003633;
    assert hash[5] = 2214481798;
    assert hash[6] = 3208491254;
    assert hash[7] = 789740750;

    return ();
}

// test vectors from: https://www.di-mgt.com.au/sha_testvectors.html

func test_sha256_0bits{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;

    let (empty) = alloc();
    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(empty, 0);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    assert hash[0] = 0xe3b0c442;
    assert hash[1] = 0x98fc1c14;
    assert hash[2] = 0x9afbf4c8;
    assert hash[3] = 0x996fb924;
    assert hash[4] = 0x27ae41e4;
    assert hash[5] = 0x649b934c;
    assert hash[6] = 0xa495991b;
    assert hash[7] = 0x7852b855;

    return ();
}

func test_sha256_24bits{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(new ('abc\x00'), 3);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    assert hash[0] = 0xba7816bf;
    assert hash[1] = 0x8f01cfea;
    assert hash[2] = 0x414140de;
    assert hash[3] = 0x5dae2223;
    assert hash[4] = 0xb00361a3;
    assert hash[5] = 0x96177a9c;
    assert hash[6] = 0xb410ff61;
    assert hash[7] = 0xf20015ad;

    return ();
}

func test_sha256_448bits{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;

    let (input) = alloc();
    assert input[0] = 'abcd';
    assert input[1] = 'bcde';
    assert input[2] = 'cdef';
    assert input[3] = 'defg';
    assert input[4] = 'efgh';
    assert input[5] = 'fghi';
    assert input[6] = 'ghij';
    assert input[7] = 'hijk';
    assert input[8] = 'ijkl';
    assert input[9] = 'jklm';
    assert input[10] = 'klmn';
    assert input[11] = 'lmno';
    assert input[12] = 'mnop';
    assert input[13] = 'nopq';

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(input, 56);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    assert hash[0] = 0x248d6a61;
    assert hash[1] = 0xd20638b8;
    assert hash[2] = 0xe5c02693;
    assert hash[3] = 0x0c3e6039;
    assert hash[4] = 0xa33ce459;
    assert hash[5] = 0x64ff2167;
    assert hash[6] = 0xf6ecedd4;
    assert hash[7] = 0x19db06c1;

    return ();
}

func test_sha256_504bits{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;
    // Input String: "0000111122223333444455556666777788889999aaaabbbbccccddddeeeefff"
    let (input) = alloc();
    assert input[0] = '0000';
    assert input[1] = '1111';
    assert input[2] = '2222';
    assert input[3] = '3333';
    assert input[4] = '4444';
    assert input[5] = '5555';
    assert input[6] = '6666';
    assert input[7] = '7777';
    assert input[8] = '8888';
    assert input[9] = '9999';
    assert input[10] = 'aaaa';
    assert input[11] = 'bbbb';
    assert input[12] = 'cccc';
    assert input[13] = 'dddd';
    assert input[14] = 'eeee';
    assert input[15] = 'fff\x00';

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(input, 63);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    // Resulting hash: 214072bf9da123ca5a8925edb05a6f071fc48fa66494d08513b9ba1b82df20cd
    assert hash[0] = 0x214072bf;
    assert hash[1] = 0x9da123ca;
    assert hash[2] = 0x5a8925ed;
    assert hash[3] = 0xb05a6f07;
    assert hash[4] = 0x1fc48fa6;
    assert hash[5] = 0x6494d085;
    assert hash[6] = 0x13b9ba1b;
    assert hash[7] = 0x82df20cd;

    return ();
}

func test_sha256_512bits{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;
    // Input String: "0000111122223333444455556666777788889999aaaabbbbccccddddeeeeffff"
    let (input) = alloc();
    assert input[0] = '0000';
    assert input[1] = '1111';
    assert input[2] = '2222';
    assert input[3] = '3333';
    assert input[4] = '4444';
    assert input[5] = '5555';
    assert input[6] = '6666';
    assert input[7] = '7777';
    assert input[8] = '8888';
    assert input[9] = '9999';
    assert input[10] = 'aaaa';
    assert input[11] = 'bbbb';
    assert input[12] = 'cccc';
    assert input[13] = 'dddd';
    assert input[14] = 'eeee';
    assert input[15] = 'ffff';

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(input, 64);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    // Resulting hash: c7a7d8c0472c7f6234380e9dd3a55eb24d3e5dba9d106b74a260dc787f2f6df8
    assert hash[0] = 0xc7a7d8c0;
    assert hash[1] = 0x472c7f62;
    assert hash[2] = 0x34380e9d;
    assert hash[3] = 0xd3a55eb2;
    assert hash[4] = 0x4d3e5dba;
    assert hash[5] = 0x9d106b74;
    assert hash[6] = 0xa260dc78;
    assert hash[7] = 0x7f2f6df8;

    return ();
}

func test_sha256_1024bits{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;
    // Input String: "0000111122223333444455556666777788889999aaaabbbbccccddddeeeeffff0000111122223333444455556666777788889999aaaabbbbccccddddeeeeffff"
    let (input) = alloc();
    assert input[0] = '0000';
    assert input[1] = '1111';
    assert input[2] = '2222';
    assert input[3] = '3333';
    assert input[4] = '4444';
    assert input[5] = '5555';
    assert input[6] = '6666';
    assert input[7] = '7777';
    assert input[8] = '8888';
    assert input[9] = '9999';
    assert input[10] = 'aaaa';
    assert input[11] = 'bbbb';
    assert input[12] = 'cccc';
    assert input[13] = 'dddd';
    assert input[14] = 'eeee';
    assert input[15] = 'ffff';
    assert input[16] = '0000';
    assert input[17] = '1111';
    assert input[18] = '2222';
    assert input[19] = '3333';
    assert input[20] = '4444';
    assert input[21] = '5555';
    assert input[22] = '6666';
    assert input[23] = '7777';
    assert input[24] = '8888';
    assert input[25] = '9999';
    assert input[26] = 'aaaa';
    assert input[27] = 'bbbb';
    assert input[28] = 'cccc';
    assert input[29] = 'dddd';
    assert input[30] = 'eeee';
    assert input[31] = 'ffff';

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(input, 128);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    // Resulting hash: e324cc62be4f0465591b5cac1309ab4d5a9ee4ae8e99158c50cef7597898f046
    assert hash[0] = 0xe324cc62;
    assert hash[1] = 0xbe4f0465;
    assert hash[2] = 0x591b5cac;
    assert hash[3] = 0x1309ab4d;
    assert hash[4] = 0x5a9ee4ae;
    assert hash[5] = 0x8e99158c;
    assert hash[6] = 0x50cef759;
    assert hash[7] = 0x7898f046;

    return ();
}

func test_sha256_896bits{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;

    let (input) = alloc();
    assert input[0] = 'abcd';
    assert input[1] = 'efgh';
    assert input[2] = 'bcde';
    assert input[3] = 'fghi';
    assert input[4] = 'cdef';
    assert input[5] = 'ghij';
    assert input[6] = 'defg';
    assert input[7] = 'hijk';
    assert input[8] = 'efgh';
    assert input[9] = 'ijkl';
    assert input[10] = 'fghi';
    assert input[11] = 'jklm';
    assert input[12] = 'ghij';
    assert input[13] = 'klmn';
    assert input[14] = 'hijk';
    assert input[15] = 'lmno';
    assert input[16] = 'ijkl';
    assert input[17] = 'mnop';
    assert input[18] = 'jklm';
    assert input[19] = 'nopq';
    assert input[20] = 'klmn';
    assert input[21] = 'opqr';
    assert input[22] = 'lmno';
    assert input[23] = 'pqrs';
    assert input[24] = 'mnop';
    assert input[25] = 'qrst';
    assert input[26] = 'nopq';
    assert input[27] = 'rstu';

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(input, 112);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    assert hash[0] = 0xcf5b16a7;
    assert hash[1] = 0x78af8380;
    assert hash[2] = 0x036ce59e;
    assert hash[3] = 0x7b049237;
    assert hash[4] = 0x0b249b11;
    assert hash[5] = 0xe8f07a51;
    assert hash[6] = 0xafac4503;
    assert hash[7] = 0x7afee9d1;

    return ();
}

func test_sha256_client_data{bitwise_ptr: BitwiseBuiltin*, range_check_ptr}() {
    alloc_locals;

    let (client_data_json) = alloc();
    assert client_data_json[0] = 2065855609;
    assert client_data_json[1] = 1885676090;
    assert client_data_json[2] = 578250082;
    assert client_data_json[3] = 1635087464;
    assert client_data_json[4] = 1848534885;
    assert client_data_json[5] = 1948396578;
    assert client_data_json[6] = 1667785068;
    assert client_data_json[7] = 1818586727;
    assert client_data_json[8] = 1696741922;
    assert client_data_json[9] = 813183028;
    assert client_data_json[10] = 879047521;
    assert client_data_json[11] = 1684224052;
    assert client_data_json[12] = 895825200;
    assert client_data_json[13] = 828518449;
    assert client_data_json[14] = 1664497968;
    assert client_data_json[15] = 878994482;
    assert client_data_json[16] = 1647338340;
    assert client_data_json[17] = 811872312;
    assert client_data_json[18] = 878862896;
    assert client_data_json[19] = 825373744;
    assert client_data_json[20] = 959854180;
    assert client_data_json[21] = 859398963;
    assert client_data_json[22] = 825636148;
    assert client_data_json[23] = 942761062;
    assert client_data_json[24] = 1667327286;
    assert client_data_json[25] = 896999980;
    assert client_data_json[26] = 577729129;
    assert client_data_json[27] = 1734962722;
    assert client_data_json[28] = 975333492;
    assert client_data_json[29] = 1953526586;
    assert client_data_json[30] = 791634799;
    assert client_data_json[31] = 1853125231;
    assert client_data_json[32] = 1819043186;
    assert client_data_json[33] = 761606451;
    assert client_data_json[34] = 1886665079;
    assert client_data_json[35] = 2004233840;
    assert client_data_json[36] = 1919252073;
    assert client_data_json[37] = 1702309475;
    assert client_data_json[38] = 1634890866;
    assert client_data_json[39] = 1768187749;
    assert client_data_json[40] = 778528546;
    assert client_data_json[41] = 740451186;
    assert client_data_json[42] = 1869837135;
    assert client_data_json[43] = 1919510377;
    assert client_data_json[44] = 1847736934;
    assert client_data_json[45] = 1634497381;
    assert client_data_json[46] = 2097152000;

    let (local sha256_ptr: felt*) = alloc();
    let sha256_ptr_start = sha256_ptr;
    let (hash) = sha256{sha256_ptr=sha256_ptr}(client_data_json, 185);
    finalize_sha256(sha256_ptr_start=sha256_ptr_start, sha256_ptr_end=sha256_ptr);

    assert hash[0] = 0x08ad1974;
    assert hash[1] = 0x216096a7;
    assert hash[2] = 0x6ff36a54;
    assert hash[3] = 0x159891a3;
    assert hash[4] = 0x57d21a90;
    assert hash[5] = 0x2c358e6f;
    assert hash[6] = 0xeb02f14c;
    assert hash[7] = 0xcaf48fcd;

    return ();
}

func main{range_check_ptr, bitwise_ptr: BitwiseBuiltin*}() {
    test_sha256_hello_world();
    test_sha256_multichunks();
    test_sha256_0bits();
    test_sha256_24bits();
    test_sha256_448bits();
    test_sha256_504bits();
    test_sha256_512bits();
    test_sha256_1024bits();
    test_sha256_896bits();
    test_sha256_client_data();
    return ();
}