// This file has been borrowed from https://github.com/mateocapon/criptografia-tp3/blob/7021d130f1af1a5904b03deeec03fe447ddcc118/euclidean_algo.cairo#L4

%builtins range_check

from starkware.cairo.common.math_cmp import is_le

func rec_integer_division_unsigned{range_check_ptr: felt}(a, b, counter) -> (felt, felt) {
    let res = is_le(a + 1, b);
    jmp end if res != 0;
    return rec_integer_division_unsigned(a - b, b, counter + 1);

    end:
    return (counter, a);
}

func integer_division_unsigned{range_check_ptr: felt}(a, b) -> (felt, felt) {
    return rec_integer_division_unsigned(a, b, 0);
}

func main{range_check_ptr: felt}() {
    let a = 16;
    let b = 5;
    let (q: felt, rem: felt) = integer_division_unsigned(a, b);
    assert q = 3;
    assert rem = 1;

    let c = 10;
    let d = 5;
    let (q2: felt, rem2: felt) = integer_division_unsigned(c, d);
    assert q2 = 2;
    assert rem2 = 0;
    return ();
}
