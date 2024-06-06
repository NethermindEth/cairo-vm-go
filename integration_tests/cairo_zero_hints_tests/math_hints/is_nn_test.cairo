// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/

%builtins range_check

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.pow import pow
from starkware.cairo.common.math import assert_le_felt, assert_lt_felt

const CONSTANT = 3 ** 10;
const RC_BOUND = 2 ** 128;

// Returns 1 if value != 0. Returns 0 otherwise.
@known_ap_change
func is_not_zero(value) -> felt {
    if (value == 0) {
        return 0;
    }

    return 1;
}

// Returns 1 if a >= 0 (or more precisely 0 <= a < RANGE_CHECK_BOUND).
// Returns 0 otherwise.
@known_ap_change
func is_nn{range_check_ptr}(a) -> felt {
    %{ memory[ap] = 0 if 0 <= (ids.a % PRIME) < range_check_builtin.bound else 1 %}
    jmp out_of_range if [ap] != 0, ap++;
    [range_check_ptr] = a;
    ap += 20;
    let range_check_ptr = range_check_ptr + 1;
    return 1;

    out_of_range:
    %{ memory[ap] = 0 if 0 <= ((-ids.a - 1) % PRIME) < range_check_builtin.bound else 1 %}
    jmp need_felt_comparison if [ap] != 0, ap++;
    assert [range_check_ptr] = (-a) - 1;
    ap += 17;
    let range_check_ptr = range_check_ptr + 1;
    return 0;

    need_felt_comparison:
    assert_le_felt(RC_BOUND, a);
    return 0;
}

// Returns 1 if a <= b (or more precisely 0 <= b - a < RANGE_CHECK_BOUND).
// Returns 0 otherwise.
@known_ap_change
func is_le{range_check_ptr}(a, b) -> felt {
    return is_nn(b - a);
}

// Returns 1 if 0 <= a <= b < RANGE_CHECK_BOUND.
// Returns 0 otherwise.
//
// Assumption: b < RANGE_CHECK_BOUND.
@known_ap_change
func is_nn_le{range_check_ptr}(a, b) -> felt {
    let res = is_nn(a);
    if (res == 0) {
        ap += 25;
        return res;
    }
    return is_nn(b - a);
}

// Returns 1 if value is in the range [lower, upper).
// Returns 0 otherwise.
// Assumptions:
//   upper - lower <= RANGE_CHECK_BOUND.
@known_ap_change
func is_in_range{range_check_ptr}(value, lower, upper) -> felt {
    let res = is_le(lower, value);
    if (res == 0) {
        ap += 26;
        return res;
    }
    return is_nn(upper - 1 - value);
}

// Checks if the unsigned integer lift (as a number in the range [0, PRIME)) of a is lower than
// or equal to that of b.
// See split_felt() for more details.
// Returns 1 if true, 0 otherwise.
@known_ap_change
func is_le_felt{range_check_ptr}(a, b) -> felt {
    %{ memory[ap] = 0 if (ids.a % PRIME) <= (ids.b % PRIME) else 1 %}
    jmp not_le if [ap] != 0, ap++;
    ap += 6;
    assert_le_felt(a, b);
    return 1;

    not_le:
    assert_lt_felt(b, a);
    return 0;
}

func fill_array_with_pow{range_check_ptr}(
    array_start: felt*, base: felt, step: felt, exp: felt, iter: felt, last: felt
) -> () {
    if (iter == last) {
        return ();
    }
    let (res) = pow(base + step, exp);
    assert array_start[iter] = res;
    return fill_array_with_pow(array_start, base + step, step, exp, iter + 1, last);
}

func test_is_not_zero{range_check_ptr}(
    base_array: felt*, new_array: felt*, iter: felt, last: felt
) -> () {
    if (iter == last) {
        return ();
    }
    let res = is_not_zero(base_array[iter]);
    assert new_array[iter] = res;
    return test_is_not_zero(base_array, new_array, iter + 1, last);
}

func test_is_nn{range_check_ptr}(base_array: felt*, new_array: felt*, iter: felt, last: felt) -> (
    ) {
    if (iter == last) {
        return ();
    }
    let res = is_nn(base_array[iter]);
    assert new_array[iter] = res;
    return test_is_nn(base_array, new_array, iter + 1, last);
}

func test_is_nn_le{range_check_ptr}(
    base_array: felt*, new_array: felt*, iter: felt, last: felt
) -> () {
    if (iter == last) {
        return ();
    }
    let res = is_nn_le(base_array[iter], CONSTANT);
    assert new_array[iter] = res;
    return test_is_nn_le(base_array, new_array, iter + 1, last);
}

func test_is_in_range{range_check_ptr}(
    base_array: felt*, new_array: felt*, iter: felt, last: felt
) -> () {
    if (iter == last) {
        return ();
    }
    let res = is_in_range(CONSTANT, base_array[iter], base_array[iter + 1]);
    assert new_array[iter] = res;
    return test_is_in_range(base_array, new_array, iter + 1, last);
}

func test_is_le_felt{range_check_ptr}(
    base_array: felt*, new_array: felt*, iter: felt, last: felt
) -> () {
    if (iter == last) {
        return ();
    }
    let res = is_le_felt(base_array[iter], CONSTANT);
    assert new_array[iter] = res;
    return test_is_le_felt(base_array, new_array, iter + 1, last);
}

func run_tests{range_check_ptr}(array_len: felt) -> () {
    alloc_locals;
    let (array: felt*) = alloc();
    fill_array_with_pow(array, 0, 3, 3, 0, array_len);

    let (array_is_not_zero: felt*) = alloc();
    test_is_not_zero(array, array_is_not_zero, 0, array_len);

    let (array_is_nn: felt*) = alloc();
    test_is_nn(array, array_is_nn, 0, array_len);

    let (array_is_le: felt*) = alloc();
    test_is_le(array, array_is_le, 0, array_len);

    let (array_is_nn_le: felt*) = alloc();
    test_is_nn_le(array, array_is_nn_le, 0, array_len);

    let (array_is_in_range: felt*) = alloc();
    test_is_in_range(array, array_is_in_range, 0, array_len - 1);

    let (array_is_le_felt: felt*) = alloc();
    test_is_le_felt(array, array_is_le_felt, 0, array_len);

    return ();
}

func main{range_check_ptr}() {
    run_tests(10);
    return ();
}