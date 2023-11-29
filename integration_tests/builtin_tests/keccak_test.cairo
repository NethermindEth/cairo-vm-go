%builtins output keccak 
from starkware.cairo.common.cairo_builtins import KeccakBuiltin
from starkware.cairo.common.keccak_state import KeccakBuiltinState

func main{output_ptr: felt*, keccak_ptr: KeccakBuiltin*}() {
    assert keccak_ptr[0].input = KeccakBuiltinState(68, 21, 35, 74, 43, 13, 31, 37);
    let result = keccak_ptr[0].output;
    let keccak_ptr = keccak_ptr + KeccakBuiltin.SIZE;
    assert [output_ptr] = result.s0;
    assert [output_ptr+1] = result.s1;
    assert [output_ptr+2] = result.s2;
    assert [output_ptr+3] = result.s3;
    assert [output_ptr+4] = result.s4;
    assert [output_ptr+5] = result.s5;
    assert [output_ptr+6] = result.s6;
    assert [output_ptr+7] = result.s7;
    let output_ptr = output_ptr + 8;
    return ();
}
