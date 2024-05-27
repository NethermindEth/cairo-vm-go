%builtins output pedersen range_check ecdsa
from starkware.cairo.common.math import assert_le_felt

func main{output_ptr: felt, pedersen_ptr: felt, range_check_ptr: felt, ecdsa_ptr: felt}() {
    assert_le_felt(1, 2);
    assert_le_felt(-2, -1);
    assert_le_felt(2, -1);
    return ();
}