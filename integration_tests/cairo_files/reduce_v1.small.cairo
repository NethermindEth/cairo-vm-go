// Receives an unreduced number, and returns a number that is equal to the original number mod
// SECP_P and in reduced form.

%builtins range_check
from starkware.cairo.common.cairo_secp.bigint3 import BigInt3
from starkware.cairo.common.cairo_secp.bigint import nondet_bigint3
from starkware.cairo.common.cairo_secp.field import UnreducedBigInt3, reduce

func main{range_check_ptr: felt}() {

    // Test one
    let x: UnreducedBigInt3 = UnreducedBigInt3(132181232131231239112312312313213083892150, 10, 10);
    let (y: BigInt3) = reduce(x);
    assert y = BigInt3(48537904510172037887998390, 1708402383786350, 10);

    // Test two
    let n: BigInt3 = reduce(UnreducedBigInt3(1321812083892150, 11230, 103321));
    assert n = BigInt3(1321812083892150, 11230, 103321);

    // Test three
    let p: BigInt3 = reduce(UnreducedBigInt3(0, 0, 0));
    assert p = BigInt3(0, 0, 0);

    // Test four
    let q: BigInt3 = reduce(UnreducedBigInt3(-10, 0, 0));
    assert q = BigInt3(
        77371252455336262886226981, 77371252455336267181195263, 19342813113834066795298815
    );

    // Test five
    let r: BigInt3 = reduce(UnreducedBigInt3(-10, -56, -111));
    assert r = BigInt3(
        77371252455336262886226981, 77371252455336267181195207, 19342813113834066795298704
    );

    return ();
}
