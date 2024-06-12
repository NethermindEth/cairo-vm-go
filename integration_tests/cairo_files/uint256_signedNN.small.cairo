// Returns 1 if the signed integer is nonnegative.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_signed_nn

func main{range_check_ptr}() {
    alloc_locals;

    // Test one
    local a: Uint256;
    a.low = 0;
    a.high = 2**127;
    uint256_signed_nn(a);
    [ap - 1] = 0;

    // Test two
    local b: Uint256;
    b.low = 0;
    b.high = 1;
    uint256_signed_nn(b);
    [ap - 1] = 1;

    return();
}
