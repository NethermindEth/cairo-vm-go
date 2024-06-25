// Unsigned integer division between two integers. Returns the quotient and the remainder.
// Conforms to EVM specifications: division by 0 yields 0.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_unsigned_div_rem

func main{range_check_ptr}() {

    // Test one
    let (quotient_one, remainder_one) = uint256_unsigned_div_rem(
        Uint256(340282366920938463454053728725133866491,0),
        Uint256(2447157533618445569039501,0)
    );
    assert quotient_one = Uint256(139052088901602,0);
    assert remainder_one = Uint256(1285139305198259893685889,0);

    // Test two
    let (quotient_two, remainder_two) = uint256_unsigned_div_rem(
        Uint256(2 ** 127, 0), Uint256(2 ** 127, 0)
    );
    assert quotient_two = Uint256(1,0);
    assert remainder_two = Uint256(0,0);

    // Test three
    let (quotient_three, remainder_three) = uint256_unsigned_div_rem(
        Uint256(5, 2 ** 127), Uint256(0, 2 ** 127)
    );
    assert quotient_three = Uint256(1, 0);
    assert remainder_three = Uint256(5, 0);

    return();
}
