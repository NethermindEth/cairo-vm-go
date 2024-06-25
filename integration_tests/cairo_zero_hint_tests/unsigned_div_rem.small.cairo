%builtins range_check

from starkware.cairo.common.math import unsigned_div_rem

func main{range_check_ptr: felt}() {
    let (q1, r1) = unsigned_div_rem(100, 5);
    assert q1 = 20;
    assert r1 = 0;

    let (q2, r2) = unsigned_div_rem(100, 21);
    assert q2 = 4;
    assert r2 = 16;

    let (q3, r3) = unsigned_div_rem(0, 1);
    assert q3 = 0;
    assert r3 = 0;

    // 2**128
    let (q4, r4) = unsigned_div_rem(340282366920938463463374607431768211456, 1234567);
    assert q4 = 275628918415070598406870269035028;
    assert r4 = 798580;

    return ();
}
