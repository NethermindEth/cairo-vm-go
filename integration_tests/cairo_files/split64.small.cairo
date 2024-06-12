// Splits a field element in the range [0, 2^192) to its low 64-bit and high 128-bit parts.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, split_64

func main{range_check_ptr}() {

    // Test one
    let (a, b) = split_64(8746);
    assert a = 8746;
    assert b = 0;

    // Test two
    let (c,d) = split_64(2**127);
    assert c = 0;
    assert d = 2 ** 63;

    return ();
}
