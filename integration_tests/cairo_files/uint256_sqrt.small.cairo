// Returns the floor value of the square root of a uint256 integer.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_sqrt

func main{range_check_ptr}() {
    alloc_locals;

    // Test one
    local a: Uint256;
    a.low = 0;
    a.high = 0;
    local sqrt_a: Uint256;
    let (sqrt_a) = uint256_sqrt(a);
    assert sqrt_a.low = 0;
    assert sqrt_a.high = 0;

    // Test two
    local b: Uint256;
    b.low = 5;
    b.high = 0;
    local sqrt_b: Uint256;
    let (sqrt_b) = uint256_sqrt(b);
    assert sqrt_b.low = 2;
    assert sqrt_b.high = 0;

    // Test three
    local c: Uint256;
    c.low = 65536;
    c.high = 0;
    local sqrt_c: Uint256;
    let (sqrt_c) = uint256_sqrt(c);
    assert sqrt_c.low = 256;
    assert sqrt_c.high = 0;

    // Test four
    local d: Uint256;
    d.low = 4294967296;
    d.high = 0;
    local sqrt_d: Uint256;
    let (sqrt_d) = uint256_sqrt(d);
    assert sqrt_d.low = 65536;
    assert sqrt_d.high = 0;

    // Test five
    local e: Uint256;
    e.low = 2**127;
    e.high = 2**127;
    local sqrt_e: Uint256;
    let (sqrt_e) = uint256_sqrt(e);
    assert sqrt_e.low = 240615969168004511545033772477625056927;
    assert sqrt_e.high = 0;

    return();
}
