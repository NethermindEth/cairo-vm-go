// The content of this file has been borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/usort.cairo

%builtins range_check
from starkware.cairo.common.usort import usort
from starkware.cairo.common.alloc import alloc

func main{range_check_ptr}() -> () {
    alloc_locals;
    let (input_array: felt*) = alloc();
    assert input_array[0] = 2;
    assert input_array[1] = 1;
    assert input_array[2] = 0;

    let (output_len, output, multiplicities) = usort(input_len=3, input=input_array);

    assert output_len = 3;
    assert output[0] = 0;
    assert output[1] = 1;
    assert output[2] = 2;
    assert multiplicities[0] = 1;
    assert multiplicities[1] = 1;
    assert multiplicities[2] = 1;
    return ();
}
