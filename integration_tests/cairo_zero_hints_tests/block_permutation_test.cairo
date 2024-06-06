// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/aecbb3f01dacb6d3f90256c808466c2c37606252/cairo_programs/keccak_alternative_hint.cairo#L20

%builtins output range_check bitwise

from starkware.cairo.common.cairo_keccak.keccak import (
    _prepare_block,
    KECCAK_FULL_RATE_IN_BYTES,
    KECCAK_FULL_RATE_IN_WORDS,
    KECCAK_STATE_SIZE_FELTS,
)
from starkware.cairo.common.math import assert_nn_le
from starkware.cairo.common.cairo_builtins import BitwiseBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.serialize import serialize_word

func _block_permutation_cairo_keccak{output_ptr: felt*, keccak_ptr: felt*}() {
    alloc_locals;
    let output = output_ptr;
    let keccak_ptr_start = keccak_ptr - KECCAK_STATE_SIZE_FELTS;
    %{
        from starkware.cairo.common.cairo_keccak.keccak_utils import keccak_func
        _keccak_state_size_felts = int(ids.KECCAK_STATE_SIZE_FELTS)
        assert 0 <= _keccak_state_size_felts < 100
        output_values = keccak_func(memory.get_range(
            ids.keccak_ptr_start, _keccak_state_size_felts))
        segments.write_arg(ids.output, output_values)
    %}
    let keccak_ptr = keccak_ptr + KECCAK_STATE_SIZE_FELTS;

    return ();
}

func run_cairo_keccak{output_ptr: felt*, range_check_ptr, bitwise_ptr: BitwiseBuiltin*}() {
    alloc_locals;

    let (output: felt*) = alloc();
    let keccak_output = output;

    let (inputs: felt*) = alloc();
    let inputs_start = inputs;
    fill_array(inputs, 9, 3, 0);

    let (state: felt*) = alloc();
    let state_start = state;
    fill_array(state, 5, 25, 0);

    let n_bytes = 24;

    _prepare_block{keccak_ptr=output_ptr}(inputs=inputs, n_bytes=n_bytes, state=state);
    _block_permutation_cairo_keccak{keccak_ptr=output_ptr}();

    local full_word: felt;
    %{ ids.full_word = int(ids.n_bytes >= 8) %}
    assert full_word = 1;

    let n_bytes = 8;
    local full_word: felt;
    %{ ids.full_word = int(ids.n_bytes >= 8) %}
    assert full_word = 1;

    let n_bytes = 7;
    local full_word: felt;
    %{ ids.full_word = int(ids.n_bytes >= 8) %}
    assert full_word = 0;

    return ();
}

func fill_array(array: felt*, base: felt, array_length: felt, iterator: felt) {
    if (iterator == array_length) {
        return ();
    }

    assert array[iterator] = base;

    return fill_array(array, base, array_length, iterator + 1);
}

func main{output_ptr: felt*, range_check_ptr, bitwise_ptr: BitwiseBuiltin*}() {
    run_cairo_keccak();

    return ();
}
