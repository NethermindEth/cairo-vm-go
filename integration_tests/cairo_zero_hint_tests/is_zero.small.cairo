// Returns 1 if x == 0 (mod secp256k1_prime), and 0 otherwise.
// Serves as integration test for the following hints : 
// isZeroNondetCode
// isZeroPackCode
// isZeroDivModCode

%builtins range_check

from starkware.cairo.common.cairo_secp.field import is_zero
from starkware.cairo.common.cairo_secp.bigint3 import SumBigInt3

func main{range_check_ptr}() -> () {

    // Test One
    let a = SumBigInt3(0, 0, 0);
    let (res: felt) = is_zero(a);
    assert res = 1;

    // Test Two
    let b = SumBigInt3(42, 0, 0);
    let (res: felt) = is_zero(b);
    assert res = 0;

    // Test Three
    let c = SumBigInt3(
        77371252455336262886226991, 77371252455336267181195263, 19342813113834066795298815
    );
    let (res: felt) = is_zero(c);
    assert res = 1;

    return ();
}
