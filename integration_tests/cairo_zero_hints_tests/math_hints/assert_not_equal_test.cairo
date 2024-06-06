// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/

from starkware.cairo.common.alloc import alloc

// Verifies that a != b. The proof will fail otherwise.
func assert_not_equal(a, b) {
    %{
        from starkware.cairo.lang.vm.relocatable import RelocatableValue
        both_ints = isinstance(ids.a, int) and isinstance(ids.b, int)
        both_relocatable = (
            isinstance(ids.a, RelocatableValue) and isinstance(ids.b, RelocatableValue) and
            ids.a.segment_index == ids.b.segment_index)
        assert both_ints or both_relocatable, \
            f'assert_not_equal failed: non-comparable values: {ids.a}, {ids.b}.'
        assert (ids.a - ids.b) % PRIME != 0, f'assert_not_equal failed: {ids.a} = {ids.b}.'
    %}
    if (a == b) {
        // If a == b, add an unsatisfiable requirement.
        a = a + 1;
    }

    return ();
}

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