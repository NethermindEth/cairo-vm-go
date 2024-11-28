%builtins output
from starkware.cairo.common.serialize import serialize_word

func main{output_ptr: felt*}() {
    alloc_locals;
    tempvar quad_pyramid_slope_angles: felt* = new (51, 52, 51, 52);
    local quad_pyramid_slope_angles: felt* = quad_pyramid_slope_angles;
    assert quad_pyramid_slope_angles[0] = 51;
    assert quad_pyramid_slope_angles[1] = 52;
    assert quad_pyramid_slope_angles[2] = 51;
    assert quad_pyramid_slope_angles[3] = 52;
    let (is_quad_valid: felt) = verify_slopes(quad_pyramid_slope_angles, 4);
    assert is_quad_valid = 1;

    tempvar tri_pyramid_slope_angles: felt* = new (51, 52, 48);
    assert tri_pyramid_slope_angles[0] = 51;
    assert tri_pyramid_slope_angles[1] = 52;
    assert tri_pyramid_slope_angles[2] = 48;
    let (is_tri_valid: felt) = verify_slopes(tri_pyramid_slope_angles, 3);
    assert is_tri_valid = 0;

    let (double_verify_res: felt) = double_verify_slopes(
        quad_pyramid_slope_angles, 4, tri_pyramid_slope_angles, 3
    );
    assert double_verify_res = 0;
    return ();
}

func verify_slopes(slopes_arr: felt*, slopes_len: felt) -> (is_valid: felt) {
    if (slopes_len == 0) {
        return (is_valid=1);
    }
    if ((slopes_arr[0] - 51) * (slopes_arr[0] - 52) == 0) {
        return verify_slopes(slopes_arr + 1, slopes_len - 1);
    }
    return (is_valid=0);
}

// do not modify code on this line or above

func double_verify_slopes(
    first_arr: felt*, first_arr_len: felt, second_arr: felt*, second_arr_len: felt
) -> (res: felt) {
    alloc_locals;
    let (local first_verify: felt) = verify_slopes(first_arr, first_arr_len);
    let (local second_verify: felt) = verify_slopes(second_arr, second_arr_len);
    if (first_verify + second_verify == 2) {
        return (res=1);
    }
    return (res=0);
}
