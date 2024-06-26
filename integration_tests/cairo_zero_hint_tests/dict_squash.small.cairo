%builtins range_check

from starkware.cairo.common.default_dict import default_dict_new
from starkware.cairo.common.dict import dict_write, dict_update, dict_squash
from starkware.cairo.common.squash_dict import squash_dict
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.dict_access import DictAccess

func test_squash_dict{range_check_ptr}() -> () {
    alloc_locals;

    let (dict_start: DictAccess*) = alloc();

    assert dict_start[0] = DictAccess(key=0, prev_value=100, new_value=100);
    assert dict_start[1] = DictAccess(key=1, prev_value=50, new_value=50);
    assert dict_start[2] = DictAccess(key=0, prev_value=100, new_value=200);
    assert dict_start[3] = DictAccess(key=1, prev_value=50, new_value=100);
    assert dict_start[4] = DictAccess(key=0, prev_value=200, new_value=300);
    assert dict_start[5] = DictAccess(key=1, prev_value=100, new_value=150);

    let dict_end = dict_start + 6 * DictAccess.SIZE;

    // (dict_start, dict_end) now represents the dictionary
    // {0: 100, 1: 50, 0: 200, 1: 100, 0: 300, 1: 150}.

    // Squash the dictionary from an array of 6 DictAccess structs
    // to an array of 2, with a single DictAccess entry per key.
    let (local squashed_dict_start: DictAccess*) = alloc();
    let (squashed_dict_end) = squash_dict{range_check_ptr=range_check_ptr}(
        dict_start, dict_end, squashed_dict_start
    );

    // Check the values of the squashed_dict
    // should be: {0: (100, 300), 1: (50, 150)}
    assert squashed_dict_start[0] = DictAccess(key=0, prev_value=100, new_value=300);
    assert squashed_dict_start[1] = DictAccess(key=1, prev_value=50, new_value=150);

    return ();
}

func test_dict_squash{range_check_ptr}() -> () {
    let (dict_start) = default_dict_new(17);
    let dict_end = dict_start;
    dict_write{dict_ptr=dict_end}(0, 1);
    dict_write{dict_ptr=dict_end}(1, 10);
    dict_write{dict_ptr=dict_end}(2, -2);
    dict_update{dict_ptr=dict_end}(0, 1, 2);
    dict_update{dict_ptr=dict_end}(0, 2, 3);
    dict_update{dict_ptr=dict_end}(0, 3, 4);
    dict_update{dict_ptr=dict_end}(1, 10, 15);
    dict_update{dict_ptr=dict_end}(1, 15, 20);
    dict_update{dict_ptr=dict_end}(1, 20, 25);
    dict_update{dict_ptr=dict_end}(2, -2, -4);
    dict_update{dict_ptr=dict_end}(2, -4, -8);
    dict_update{dict_ptr=dict_end}(2, -8, -16);
    let (squashed_dict_start, squashed_dict_end) = dict_squash{range_check_ptr=range_check_ptr}(
        dict_start, dict_end
    );
    assert squashed_dict_end[0] = DictAccess(key=0, prev_value=1, new_value=4);
    assert squashed_dict_end[1] = DictAccess(key=1, prev_value=10, new_value=25);
    assert squashed_dict_end[2] = DictAccess(key=2, prev_value=-2, new_value=-16);
    return ();
}

func main{range_check_ptr}() {
    test_squash_dict();
    test_dict_squash();

    return ();
}
