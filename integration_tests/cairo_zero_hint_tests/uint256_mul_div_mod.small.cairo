// Computes:
// 1. The integer division `(a * b) // div` (as a 512-bit number).
// 2. The remainder `(a * b) modulo div`.
// Assumption: div != 0.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_mul_div_mod

func main{range_check_ptr}() {

    // Test one
    let (quotient_one_low, quotient_one_high, remainder_one) = uint256_mul_div_mod(
        Uint256(89, 72), Uint256(3, 7), Uint256(107, 114)
    );
    assert quotient_one_low = Uint256(143276786071974089879315624181797141668, 4);
    assert quotient_one_high = Uint256(0, 0);
    assert remainder_one = Uint256(322372768661941702228460154409043568767, 101);

    // Test two
    let (quotient_two_low, quotient_two_high, remainder_two) = uint256_mul_div_mod(
        Uint256(340281070833283907490476236129005105807, 340282366920938463463374607431768211455),
        Uint256(2447157533618445569039502, 0),
        Uint256(0, 1),
    );
    assert quotient_two_low = Uint256(
        340282366920938463454053728725133866491, 2447157533618445569039501
    );
    assert quotient_two_high = Uint256(0, 0);
    assert remainder_two = Uint256(326588112914912836985603897252688572242, 0);

    // Test three
    let (quotient_three_low, quotient_three_high, remainder_three) = uint256_mul_div_mod(
        Uint256(0, 2 ** 127),
        Uint256(0, 2 ** 127),
        Uint256(2 ** 126, 0)
    );
    assert quotient_three_low = Uint256(0, 0);
    assert quotient_three_high = Uint256(0, 1);
    assert remainder_three = Uint256(0, 0);

    return();
}
