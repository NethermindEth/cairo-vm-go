// Computes a * b^(-1) modulo the size of the elliptic curve (N).

// Serves as integration test for both divModNSafeDivCode and divModNPackedDivmodV1Code

%builtins range_check

from starkware.cairo.common.cairo_secp.bigint import BigInt3
from starkware.cairo.common.cairo_secp.signature import div_mod_n

func main{range_check_ptr: felt}() {

    // Test one
    let a: BigInt3 = BigInt3(1, 0, 0);
    let b: BigInt3 = BigInt3(1, 0, 0);

    let (res_one) = div_mod_n(a, b);

    assert res_one = BigInt3(1, 0, 0);

    // Test two
    let c: BigInt3 = BigInt3(424, 467, 182);
    let d: BigInt3 = BigInt3(12, 13, 14);

    let (res_two) = div_mod_n(c, d);

    assert res_two = BigInt3(
        62377162754175619751492349, 74037994490090883982730393, 7758940960710595272998044
    );

    // Test three
    let e: BigInt3 = BigInt3(100, 99, 98);
    let f: BigInt3 = BigInt3(10, 9, 8);

    let (res_three) = div_mod_n(e, f);

    assert res_three = BigInt3(
        3413472211745629263979533, 17305268010345238170172332, 11991751872105858217578135
    );

    return ();
}
