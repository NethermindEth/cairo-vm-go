// Adds two integers. Returns the result as a 256-bit integer and the (1-bit) carry.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_add

func main{range_check_ptr}() {
    alloc_locals;

    // Test one
    local a: Uint256;
    a.low = 2**127;
    a.high = 0;
    local b: Uint256;
    b.low = 2**127;
    b.high = 0;
    local res_one: Uint256;
    let (res_one, carry_high_one) = uint256_add(a, b);
    assert res_one.low = 0;
    assert res_one.high = 1;
    assert carry_high_one = 0;

    // Test two
    local c: Uint256;
    c.low = 2**127;
    c.high = 2**127;
    local d: Uint256;
    d.low = 2**127;
    d.high = 2**127;
    local res_two: Uint256;
    let (res_two, carry_high_two) = uint256_add(c, d);
    assert res_two.low = 0;
    assert res_two.high = 1;
    assert carry_high_two = 1;

    // Test three
    local e: Uint256;
    e.low = 0;
    e.high = 2**127;
    local f: Uint256;
    f.low = 0;
    f.high = 2**127;
    local res_three: Uint256;
    let (res_three, carry_high_three) = uint256_add(e, f);
    assert res_three.low = 0;
    assert res_three.high = 0;
    assert carry_high_three = 1;

    // Test four
    local g: Uint256;
    g.low = 2**127;
    g.high = 0;
    local h: Uint256;
    h.low = 0;
    h.high = 2**127;
    local res_four: Uint256;
    let (res_four, carry_high_four) = uint256_add(g, h);
    assert res_four.low = 2**127;
    assert res_four.high = 2**127;
    assert carry_high_four = 0;

    // Test five
    local i: Uint256;
    i.low = 426942694269;
    i.high = 426942694269;
    local j: Uint256;
    j.low = 426942694269;
    j.high = 426942694269;
    local res_five: Uint256;
    let (res_five, carry_high_five) = uint256_add(i, j);
    assert res_five.low = 853885388538;
    assert res_five.high = 853885388538;
    assert carry_high_five = 0;

    return ();
}
