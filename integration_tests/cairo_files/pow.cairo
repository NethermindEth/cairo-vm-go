%builtins range_check

from starkware.cairo.common.pow import pow

func main{range_check_ptr: felt}() {
    let (x) = pow(5, 3);
    assert x = 125;
    let (x) = pow(4, 3);
    assert x = 64;
    let (x) = pow(2, 10);
    assert x = 1024;
    let (y) = pow(x, 2);
    assert x = 1048576;
    return ();
}
