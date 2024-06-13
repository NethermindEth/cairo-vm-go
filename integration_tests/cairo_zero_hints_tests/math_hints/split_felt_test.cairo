// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/split_felt.cairo

%builtins range_check

from starkware.cairo.common.math import assert_le
from starkware.cairo.common.math import split_felt

func main{range_check_ptr: felt}() {
    let (x, y) = split_felt(5784800237655953878877368326340059594760);
    assert x = 17;
    assert y = 8;
    return ();
}
