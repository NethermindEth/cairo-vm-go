%builtins range_check
from starkware.cairo.common.find_element import find_element
from starkware.cairo.common.alloc import alloc

func main{range_check_ptr}() -> () {
    alloc_locals;
    let (local array_ptr: felt*) = alloc();
    assert array_ptr[0] = 1;

    let (element_ptr: felt*) = find_element(
        array_ptr=array_ptr, elm_size=1, n_elms=1, key=1
    );
    assert [element_ptr] = 1;

    return ();
}
