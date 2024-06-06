%builtins range_check

from starkware.cairo.common.alloc import alloc

// Asserts that 'value' is in the range [0, 2**250).
@known_ap_change
func assert_250_bit{range_check_ptr}(value) {
    const UPPER_BOUND = 2 ** 250;
    const SHIFT = 2 ** 128;
    const HIGH_BOUND = UPPER_BOUND / SHIFT;

    let low = [range_check_ptr];
    let high = [range_check_ptr + 1];

    %{
        from starkware.cairo.common.math_utils import as_int

        # Correctness check.
        value = as_int(ids.value, PRIME) % PRIME
        assert value < ids.UPPER_BOUND, f'{value} is outside of the range [0, 2**250).'

        # Calculation for the assertion.
        ids.high, ids.low = divmod(ids.value, ids.SHIFT)
    %}

    assert [range_check_ptr + 2] = HIGH_BOUND - 1 - high;

    // The assert below guarantees that
    //   value = high * SHIFT + low <= (HIGH_BOUND - 1) * SHIFT + 2**128 - 1 =
    //   HIGH_BOUND * SHIFT - SHIFT + SHIFT - 1 = 2**250 - 1.
    assert value = high * SHIFT + low;

    let range_check_ptr = range_check_ptr + 3;
    return ();
}

func assert_250_bit_element_array{range_check_ptr: felt}(
    array: felt*, array_length: felt, iterator: felt
) {
    if (iterator == array_length) {
        return ();
    }
    assert_250_bit(array[iterator]);
    return assert_250_bit_element_array(array, array_length, iterator + 1);
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
    let (array: felt*) = alloc();
    fill_array(array, 70000000000000000000, 300000000000000000, array_length, 0);
    assert_250_bit_element_array(array, array_length, 0);
    return ();
}