// Returns the floor value of the square root of a uint256 integer.

%builtins range_check

from starkware.cairo.common.uint256 import Uint256, uint256_sqrt

func main{range_check_ptr}() {

    // Test one
    let (sqrt_a) = uint256_sqrt(
        Uint256(170141183460469231731687303715884105688, 0)
    );
    assert sqrt_a = Uint256(13043817825332782212, 0);

    // Test two
    let (sqrt_b) = uint256_sqrt(
        Uint256(143276786071974089879315624181797141668, 0)
    );
    assert sqrt_b = Uint256(11969828155490541121, 0);

    // Test three
    let (sqrt_c) = uint256_sqrt(
        Uint256(340282366920938463454053728725133866491, 0)
    );
    assert sqrt_c = Uint256(18446744073709551615, 0);

    // Test four
    let (sqrt_d) = uint256_sqrt(
        Uint256(2447157533618445569039501,0)
    );
    assert sqrt_d = Uint256(1564339328156,0);

    // Test five
    let (sqrt_e) = uint256_sqrt(
        Uint256(2 ** 127, 2 ** 127)
    );
    assert sqrt_e = Uint256(240615969168004511545033772477625056927, 0);

    return();
}
