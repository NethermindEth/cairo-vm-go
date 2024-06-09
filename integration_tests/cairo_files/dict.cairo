// inspired from the dict.cairo integration test in the lambdaclass cairo-vm codebase

from starkware.cairo.common.default_dict import default_dict_new
from starkware.cairo.common.dict import dict_read, dict_write
from starkware.cairo.common.dict_access import DictAccess

func test_default_dict() {
    alloc_locals;
    let (local my_dict: DictAccess*) = default_dict_new(123);

    return ();
}

func test_read() {
    alloc_locals;
    let (local my_dict: DictAccess*) = default_dict_new(123);

    let (local val1: felt) = dict_read{dict_ptr=my_dict}(key=1);
    assert val1 = 123;

    let (local val2: felt) = dict_read{dict_ptr=my_dict}(key=2);
    assert val2 = 123;

    return ();
}

func test_write() {
    alloc_locals;
    let (local my_dict: DictAccess*) = default_dict_new(123);

    let (local val1: felt) = dict_read{dict_ptr=my_dict}(key=1);
    assert val1 = 123;

    let (local val2: felt) = dict_read{dict_ptr=my_dict}(key=2);
    assert val2 = 123;

    dict_write{dict_ptr=my_dict}(key=1, new_value=512);
    let (local val3: felt) = dict_read{dict_ptr=my_dict}(key=1);
    assert val3 = 512;

    let (local val4: felt) = dict_read{dict_ptr=my_dict}(key=2);
    assert val4 = 123;

    dict_write{dict_ptr=my_dict}(key=1, new_value=1024);
    let (local val5: felt) = dict_read{dict_ptr=my_dict}(key=1);
    assert val5 = 1024;

    let (local val6: felt) = dict_read{dict_ptr=my_dict}(key=2);
    assert val6 = 123;

    dict_write{dict_ptr=my_dict}(key=1, new_value=888);
    dict_write{dict_ptr=my_dict}(key=2, new_value=999);
    let (local val7: felt) = dict_read{dict_ptr=my_dict}(key=1);
    assert val7 = 888;
    let (local val8: felt) = dict_read{dict_ptr=my_dict}(key=2);
    assert val8 = 999;
    let (local val9: felt) = dict_read{dict_ptr=my_dict}(key=3);
    assert val9 = 123;

    return ();
}

func main() {
    test_default_dict();
    test_read();
    test_write();

    return ();
}

