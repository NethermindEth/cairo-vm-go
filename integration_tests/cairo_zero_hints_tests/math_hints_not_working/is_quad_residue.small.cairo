// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/is_quad_residue_test.cairo

%builtins output
from starkware.cairo.common.serialize import serialize_word
from starkware.cairo.common.math import is_quad_residue
from starkware.cairo.common.alloc import alloc

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
