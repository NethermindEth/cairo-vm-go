// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/

%builtins output
from starkware.cairo.common.serialize import serialize_word
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.bool import FALSE, TRUE

// Returns TRUE if `x` is a quadratic residue modulo the STARK prime. Returns FALSE otherwise.
// Returns TRUE on 0.
@known_ap_change
func is_quad_residue(x: felt) -> felt {
    alloc_locals;
    local y;
    %{
        from starkware.crypto.signature.signature import FIELD_PRIME
        from starkware.python.math_utils import div_mod, is_quad_residue, sqrt

        x = ids.x
        if is_quad_residue(x, FIELD_PRIME):
            ids.y = sqrt(x, FIELD_PRIME)
        else:
            ids.y = sqrt(div_mod(x, 3, FIELD_PRIME), FIELD_PRIME)
    %}
    // Relies on the fact that 3 is not a quadratic residue modulo the prime, so for every field
    // element x, either:
    //   * x is a quadratic residue and there exists y such that y^2 = x.
    //   * x is not a quadratic residue and there exists y such that 3 * y^2 = x.
    tempvar y_squared = y * y;
    if (y_squared == x) {
        ap += 1;
        return TRUE;
    } else {
        assert 3 * y_squared = x;
        return FALSE;
    }
}

func fill_array(array_start: felt*, iter: felt) -> () {
    if (iter == 10) {
        return ();
    }
    assert array_start[iter] = iter;
    return fill_array(array_start, iter + 1);
}

func check_quad_res{output_ptr: felt*}(inputs: felt*, expected: felt*, iter: felt) {
    if (iter == 10) {
        return ();
    }
    serialize_word(inputs[iter]);
    serialize_word(expected[iter]);

    assert is_quad_residue(inputs[iter]) = expected[iter];
    return check_quad_res(inputs, expected, iter + 1);
}

func main{output_ptr: felt*}() {
    alloc_locals;
    let (inputs: felt*) = alloc();
    fill_array(inputs, 0);

    let (expected: felt*) = alloc();
    assert expected[0] = 1;
    assert expected[1] = 1;
    assert expected[2] = 1;
    assert expected[3] = 0;
    assert expected[4] = 1;
    assert expected[5] = 1;
    assert expected[6] = 0;
    assert expected[7] = 1;
    assert expected[8] = 1;
    assert expected[9] = 1;

    check_quad_res(inputs, expected, 0);

    return ();
}