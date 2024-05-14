from starkware.cairo.common.default_dict import default_dict_new
from starkware.cairo.common.dict_access import DictAccess

func test_default_dict() {
    alloc_locals;
    let (local my_dict: DictAccess*) = default_dict_new(123);

    return ();
}

func main() {
    test_default_dict();

    return ();
}
