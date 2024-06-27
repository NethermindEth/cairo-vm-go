// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/signed_div_rem.cairo

%builtins output range_check
from starkware.cairo.common.math import signed_div_rem, assert_le
from starkware.cairo.common.serialize import serialize_word

func main{output_ptr: felt*, range_check_ptr: felt}() {
    let (q_negative, r_negative) = signed_div_rem(-10, 3, 29);
    let (q, r) = signed_div_rem(10, 3, 29);

    return ();
}
