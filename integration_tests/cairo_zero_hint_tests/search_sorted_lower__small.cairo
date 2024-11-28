// The content of this file has been borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/5d1181185a976c77956aaa4247846babd4d0e2df/cairo_programs/search_sorted_lower.cairo

%builtins range_check
from starkware.cairo.common.find_element import search_sorted_lower
from starkware.cairo.common.alloc import alloc

struct MyStruct {
    a: felt,
    b: felt,
}

func main{range_check_ptr}() -> () {
    // Create an array with MyStruct elements (1,2), (3,4), (5,6).
    alloc_locals;
    let (local array_ptr: MyStruct*) = alloc();
    assert array_ptr[0] = MyStruct(a=1, b=2);
    assert array_ptr[1] = MyStruct(a=3, b=4);
    assert array_ptr[2] = MyStruct(a=3, b=8);
    assert array_ptr[3] = MyStruct(a=5, b=6);

    let (smallest_ptr: MyStruct*) = search_sorted_lower(
        array_ptr=array_ptr, elm_size=2, n_elms=3, key=2
    );
    assert smallest_ptr.a = 3;
    assert smallest_ptr.b = 4;

    let (local array_ptr: MyStruct*) = alloc();
    assert array_ptr[0] = MyStruct(a=1, b=2);
    assert array_ptr[1] = MyStruct(a=3, b=4);
    assert array_ptr[2] = MyStruct(a=3, b=8);
    assert array_ptr[3] = MyStruct(a=5, b=6);

    let (smallest_ptr: MyStruct*) = search_sorted_lower(
        array_ptr=array_ptr, elm_size=2, n_elms=4, key=6
    );
    assert smallest_ptr.a = 0;
    assert smallest_ptr.b = 0;
    
    return ();
}
