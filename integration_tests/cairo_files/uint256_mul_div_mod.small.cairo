// Computes:
// 1. The integer division `(a * b) // div` (as a 512-bit number).
// 2. The remainder `(a * b) modulo div`.
// Assumption: div != 0.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_mul_div_mod

func main{range_check_ptr}() {
    alloc_locals;

    // Test one
    local a_one: Uint256;
    a_one.low = 6;
    a_one.high = 0;
    local b_one: Uint256;
    b_one.low = 6;
    b_one.high = 0;
    local div_one: Uint256;
    div_one.low = 2;
    div_one.high = 0;
    local quotient_one_low: Uint256;
    local quotient_one_high: Uint256;
    local remainder_one: Uint256;
    let (quotient_one_low, quotient_one_high, remainder_one) = uint256_mul_div_mod(a_one, b_one, div_one);
    assert quotient_one_low.low = 18;
    assert quotient_one_low.high = 0;
    assert quotient_one_high.low = 0;
    assert quotient_one_high.high = 0;
    assert remainder_one.low = 0;
    assert remainder_one.high = 0;

    // Test two
    local a_two: Uint256;
    a_two.low = 0;
    a_two.high = 2;
    local b_two: Uint256;
    b_two.low = 0;
    b_two.high = 3;
    local div_two: Uint256;
    div_two.low = 0;
    div_two.high = 2;
    local quotient_two_low: Uint256;
    local quotient_two_high: Uint256;
    local remainder_two: Uint256;
    let (quotient_two_low, quotient_two_high, remainder_two) = uint256_mul_div_mod(a_two, b_two, div_two);
    assert quotient_two_low.low = 0;
    assert quotient_two_low.high = 3;
    assert quotient_two_high.low = 0;
    assert quotient_two_high.high = 0;
    assert remainder_two.low = 0;
    assert remainder_two.high = 0;

    return();
}
