// Adds two integers. Returns the result as a 256-bit integer and the (1-bit) carry.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_add

func main{range_check_ptr}() {
    // Test one
    let (res_one, carry_high_one) = uint256_add(
        Uint256(2 ** 127, 0), Uint256(2 ** 127, 0)
    );
    assert res_one = Uint256(0, 1);
    assert carry_high_one = 0;

    // Test two
    let (res_two, carry_high_two) = uint256_add(
        Uint256(2 ** 127, 2 ** 127), Uint256(2 ** 127, 2 ** 127)
    );
    assert res_two = Uint256(0, 1);
    assert carry_high_two = 1;

    // Test three
    let (res_three, carry_high_three) = uint256_add(
        Uint256(0, 2 ** 127), Uint256(0, 2 ** 127)
    );
    assert res_three = Uint256(0, 0);
    assert carry_high_three = 1;

    // Test four
    let (res_four, carry_high_four) = uint256_add(
        Uint256(2 ** 127, 0), Uint256(0, 2 ** 127)
    );
    assert res_four = Uint256(2 ** 127, 2 ** 127);
    assert carry_high_four = 0;

    // Test five
    let (res_five, carry_high_five) = uint256_add(
        Uint256(426942694269 , 426942694269), Uint256(426942694269 , 426942694269)
    );
    assert res_five = Uint256(853885388538, 853885388538);
    assert carry_high_five = 0;

    return ();
}
