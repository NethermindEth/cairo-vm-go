// Returns 1 if the signed integer is nonnegative.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_signed_nn

func main{range_check_ptr}() {

    // Test one
    uint256_signed_nn(
        Uint256(0, 2 ** 127)
    );
    [ap - 1] = 0;

    // Test two
    uint256_signed_nn(
        Uint256(0, 1)
    );
    [ap - 1] = 1;

    return();
}
