%builtins output range_check bitwise

from starkware.cairo.common.cairo_keccak.keccak import _keccak
from starkware.cairo.common.cairo_builtins import BitwiseBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.serialize import serialize_word

func fill_array(array: felt*, base: felt, array_length: felt, iterator: felt) {
    if (iterator == array_length) {
        return ();
    }

    assert array[iterator] = base;

    return fill_array(array, base, array_length, iterator + 1);
}

func main{output_ptr: felt*, range_check_ptr, bitwise_ptr: BitwiseBuiltin*}() {
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

    let (res: felt*) = _keccak{keccak_ptr=keccak_output}(
        inputs=inputs_start, n_bytes=n_bytes, state=state_start
    );

    serialize_word(res[0]);
    serialize_word(res[1]);
    serialize_word(res[2]);
    serialize_word(res[4]);

    return ();
}
