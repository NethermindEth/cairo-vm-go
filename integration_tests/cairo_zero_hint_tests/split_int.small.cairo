%builtins range_check

from starkware.cairo.common.math import split_int
from starkware.cairo.common.alloc import alloc

func main{range_check_ptr: felt}() {
    alloc_locals;
    let value = 9876543210;
    let n = 10;
    let base = 10;
    let bound = 100;
    let output: felt* = alloc();
    split_int(value, n, base, bound, output);
    assert output[0] = 0;
    assert output[1] = 1;
    assert output[2] = 2;
    assert output[3] = 3;
    assert output[4] = 4;
    assert output[5] = 5;
    assert output[6] = 6;
    assert output[7] = 7;
    assert output[8] = 8;
    assert output[9] = 9;
    return ();
}
