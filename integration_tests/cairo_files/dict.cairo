from starkware.cairo.common.default_dict import default_dict_new
from starkware.cairo.common.dict import dict_read, dict_write, dict_update
from starkware.cairo.common.dict_access import DictAccess

func main() {
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

    dict_update{dict_ptr=my_dict}(key=1, prev_value=1024, new_value=2048);
    let (local val7: felt) = dict_read{dict_ptr=my_dict}(key=1);
    assert val7 = 2048;

    return ();
}
