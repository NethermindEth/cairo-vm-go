%builtins range_check

from starkware.cairo.common.pow import pow

func main{range_check_ptr: felt}() {
    let (x) = pow(5, 3);
    assert x = 125;
    let (x) = pow(4, 3);
    assert x = 64;
    return ();
}
