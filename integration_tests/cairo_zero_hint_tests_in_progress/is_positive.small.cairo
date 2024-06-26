// The content of this file has been borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/abs_value_array.cairo

%builtins range_check

from starkware.cairo.common.bool import TRUE, FALSE
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.math import abs_value

func compare_abs_arrays{range_check_ptr: felt}(
    array_a: felt*, array_b: felt*, array_length: felt, iterator: felt
) -> (r: felt) {
    if (iterator == array_length) {
        return (TRUE,);
    }
    let abs_a = abs_value(array_a[iterator]);
    let abs_b = abs_value(array_b[iterator]);
    if (abs_a != abs_b) {
        return (FALSE,);
    }
    return compare_abs_arrays(array_a, array_b, array_length, iterator + 1);
}

func fill_array(array: felt*, base: felt, step: felt, array_length: felt, iterator: felt) {
    if (iterator == array_length) {
        return ();
    }
    assert array[iterator] = base + step * iterator;
    return fill_array(array, base, step, array_length, iterator + 1);
}

func main{range_check_ptr: felt}() {
    alloc_locals;
    tempvar array_length = 10;
    let (array_a: felt*) = alloc();

    let (array_b: felt*) = alloc();
    fill_array(array_a, -7, -3, array_length, 0);
    fill_array(array_b, 7, 3, array_length, 0);
    let result: felt = compare_abs_arrays(array_a, array_b, array_length, 0);
    assert result = TRUE;
    return ();
}
