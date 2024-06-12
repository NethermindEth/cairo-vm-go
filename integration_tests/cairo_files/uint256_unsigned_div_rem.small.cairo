// Unsigned integer division between two integers. Returns the quotient and the remainder.
// Conforms to EVM specifications: division by 0 yields 0.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_unsigned_div_rem

func main{range_check_ptr}() {
    alloc_locals;

    // Test one
    local a_one: Uint256;
    a_one.low = 6;
    a_one.high = 0;
    local div_one: Uint256;
    div_one.low = 2;
    div_one.high = 0;
    local quotient_one: Uint256;
    local remainder_one: Uint256;
    let (quotient_one, remainder_one) = uint256_unsigned_div_rem(a_one, div_one);
    assert quotient_one.low = 3;
    assert quotient_one.high = 0;
    assert remainder_one.low = 0;
    assert remainder_one.high = 0;

    // Test two
    local a_two: Uint256;
    a_two.low = 2**127;
    a_two.high = 0;
    local div_two: Uint256;
    div_two.low = 2**127;
    div_two.high = 0;
    local quotient_two: Uint256;
    local remainder_two: Uint256;
    let (quotient_two, remainder_two) = uint256_unsigned_div_rem(a_two, div_two);
    assert quotient_two.low = 1;
    assert quotient_two.high = 0;
    assert remainder_two.low = 0;
    assert remainder_two.high = 0;

    // Test three
    local a_three: Uint256;
    a_three.low = 5;
    a_three.high = 2**127;
    local div_three: Uint256;
    div_three.low = 0;
    div_three.high = 2**127;
    local quotient_three: Uint256;
    local remainder_three: Uint256;
    let (quotient_three, remainder_three) = uint256_unsigned_div_rem(a_three, div_three);
    assert quotient_three.low = 1;
    assert quotient_three.high = 0;
    assert remainder_three.low = 5;
    assert remainder_three.high = 0;

    return();
}
