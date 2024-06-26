// The content of this file has been borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/compare_different_arrays.cairo

from starkware.cairo.common.math import assert_not_equal
from starkware.cairo.common.alloc import alloc

func compare_different_arrays(array_a: felt*, array_b: felt*, array_length: felt, iterator: felt) {
    if (iterator == array_length) {
        return ();
    }
    assert_not_equal(array_a[iterator], array_b[iterator]);
    return compare_different_arrays(array_a, array_b, array_length, iterator + 1);
}

func fill_array(array: felt*, base: felt, step: felt, array_length: felt, iterator: felt) {
    if (iterator == array_length) {
        return ();
    }
    assert array[iterator] = base + step * iterator;
    return fill_array(array, base, step, array_length, iterator + 1);
}

func main() {
    alloc_locals;
    tempvar array_length = 10;
    let (array_a: felt*) = alloc();
    let (array_b: felt*) = alloc();
    fill_array(array_a, 3, 90, array_length, 0);
    fill_array(array_b, 7, 3, array_length, 0);
    compare_different_arrays(array_a, array_b, array_length, 0);
    return ();
}
