// inspired from the dict.cairo integration test in the lambdaclass cairo-vm codebase

from starkware.cairo.common.default_dict import default_dict_new
from starkware.cairo.common.dict import dict_read
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

func main() {
    test_default_dict();
    test_read();

    return ();
}
