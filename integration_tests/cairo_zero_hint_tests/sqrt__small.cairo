%builtins range_check

from starkware.cairo.common.math import sqrt

func main{range_check_ptr: felt}() {
    let val1 = sqrt(0);
    assert val1 = 0;

    let val2 = sqrt(1);
    assert val2 = 1;

    let val3 = sqrt(1024);
    assert val3 = 32;

    let val4 = sqrt(99999);
    assert val4 = 316;

    // 2**250 - 1
    let val5 = sqrt(1809251394333065553493296640760748560207343510400633813116524750123642650623);
    assert val5 = 42535295865117307932921825928971026431;

    return ();
}
