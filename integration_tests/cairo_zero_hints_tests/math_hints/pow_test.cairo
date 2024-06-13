// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/pow.cairo

%builtins range_check

from starkware.cairo.common.pow import pow
from starkware.cairo.common.registers import get_ap, get_fp_and_pc

func main{range_check_ptr: felt}() {
    let (x) = pow(2, 3);
    assert x = 8;
    let (y) = pow(10, 6);
    assert y = 1000000;
    let (z) = pow(152, 25);
    assert z = 3516330588649452857943715400722794159857838650852114432;
    let (u) = pow(-2, 3);
    assert (u) = -8;
    let (v) = pow(-25, 31);
    assert (v) = -21684043449710088680149056017398834228515625;

    return ();
}
