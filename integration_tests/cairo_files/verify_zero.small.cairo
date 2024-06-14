// Verifies that the given unreduced value is equal to zero modulo the secp256k1 prime.

%builtins range_check

from starkware.cairo.common.cairo_secp.bigint import UnreducedBigInt3
from starkware.cairo.common.cairo_secp.field import verify_zero

func main{range_check_ptr}() {

    // Test one
    let a = UnreducedBigInt3(0,0,0);
    verify_zero(a);

    // Test two
    let b = UnreducedBigInt3(77371252455336262886226991,77371252455336267181195263,19342813113834066795298815);
    verify_zero(b);

    // Test three
    let c = UnreducedBigInt3(77371252455336258591258718, 77371252455336267181195263, 38685626227668133590597631);
    verify_zero(c);

    // Test four
    let d = UnreducedBigInt3(77371252455336254296290445, 77371252455336267181195263, 58028439341502200385896447);
    verify_zero(d);

    // Test five
    let e = UnreducedBigInt3(77371252455336250001322172, 77371252455336267181195263, 77371252455336267181195263);
    verify_zero(e);

    return();
}
