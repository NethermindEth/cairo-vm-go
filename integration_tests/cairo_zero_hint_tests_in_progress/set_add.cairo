%builtins range_check

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.set import set_add

struct MyStruct {
    a: felt,
    b: felt,
}

func main{range_check_ptr}() {
    alloc_locals;

    // An array containing two structs.
    let (local my_list: MyStruct*) = alloc();
    assert my_list[0] = MyStruct(a=1, b=3);
    assert my_list[1] = MyStruct(a=5, b=7);

    // Suppose that we want to add the element
    // MyStruct(a=2, b=3) to my_list, but only if it is not already
    // present (for the purpose of the example the contents of the
    // array are known, but this doesn't have to be the case)
    let list_end: felt* = &my_list[2];
    let (new_elm: MyStruct*) = alloc();
    assert new_elm[0] = MyStruct(a=2, b=3);

    set_add{set_end_ptr=list_end}(set_ptr=my_list, elm_size=MyStruct.SIZE, elm_ptr=new_elm);
    assert my_list[2] = MyStruct(a=2, b=3);
    return ();
}
