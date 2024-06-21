%builtins range_check
from starkware.cairo.common.find_element import find_element
from starkware.cairo.common.alloc import alloc

struct MyStruct {
    a: felt,
    b: felt,
}

func find_element_with_generic_key{range_check_ptr}() -> () {
    alloc_locals;
    let (local array_ptr: felt*) = alloc();
    assert array_ptr[0] = 1;

    let (element_ptr: felt*) = find_element(
        array_ptr=array_ptr, elm_size=1, n_elms=1, key=1
    );
    assert [element_ptr] = 1;

    return ();
}

struct TestStruct {
    a: felt,
    b: felt,
    c: felt,
}

func find_element_with_struct_key{range_check_ptr}() -> () {
    alloc_locals;
    let (local array_ptr: TestStruct*) = alloc();
    assert array_ptr[0] = TestStruct(a=111, b=112, c=113);
    assert array_ptr[1] = TestStruct(a=211, b=212, c=213);
    assert array_ptr[2] = TestStruct(a=311, b=312, c=313);
    assert array_ptr[3] = TestStruct(a=411, b=412, c=413);
    assert array_ptr[4] = TestStruct(a=511, b=512, c=513);
    assert array_ptr[5] = TestStruct(a=611, b=612, c=613);

    let (element_ptr: TestStruct*) = find_element(
        array_ptr=array_ptr, elm_size=TestStruct.SIZE, n_elms=6, key=311
    );

    assert element_ptr.a = 311;
    assert element_ptr.b = 312;
    assert element_ptr.c = 313;

    return ();
}

func main{range_check_ptr}() -> () {
    // Create an array with MyStruct elements (1,2), (3,4), (5,6).
    alloc_locals;
    let (local array_ptr: MyStruct*) = alloc();
    assert array_ptr[0] = MyStruct(a=1, b=2);
    assert array_ptr[1] = MyStruct(a=3, b=4);
    assert array_ptr[2] = MyStruct(a=5, b=6);

    // Find any element with key '5'.
    let (element_ptr: MyStruct*) = find_element(
        array_ptr=array_ptr, elm_size=MyStruct.SIZE, n_elms=3, key=5
    );
    // A pointer to the element with index 2 is returned.
    assert element_ptr.a = 5;
    assert element_ptr.b = 6;

    find_element_with_generic_key();
    find_element_with_struct_key();

    return ();
}