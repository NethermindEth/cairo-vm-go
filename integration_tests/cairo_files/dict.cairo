from starkware.cairo.common.default_dict import default_dict_new
from starkware.cairo.common.dict import dict_read
from starkware.cairo.common.dict_access import DictAccess

func main() {
    alloc_locals;
    let (local my_dict: DictAccess*) = default_dict_new(123);
    let (local val1: felt) = dict_read{dict_ptr=my_dict}(key=1);
    assert val1 = 123;
    let (local val2: felt) = dict_read{dict_ptr=my_dict}(key=2);
    assert val2 = 123;
    return ();
}
