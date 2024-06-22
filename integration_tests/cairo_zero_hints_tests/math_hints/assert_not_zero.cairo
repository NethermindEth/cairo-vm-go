// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/assert_not_zero.cairo

from starkware.cairo.common.math import assert_not_zero

func main() {
    assert_not_zero(1);
    assert_not_zero(-1);
    let x = 500 * 5;
    assert_not_zero(x);
    tempvar y = -80;
    assert_not_zero(y);

    return ();
}
