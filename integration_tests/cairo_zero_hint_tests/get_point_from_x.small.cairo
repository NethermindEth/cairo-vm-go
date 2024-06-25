// Returns a point on the secp256k1 curve with the given x coordinate. Chooses the y that has the
// same parity as v (there are two y values that correspond to x, with different parities).

%builtins range_check

from starkware.cairo.common.cairo_secp.signature import get_point_from_x
from starkware.cairo.common.cairo_secp.bigint import BigInt3

func main{range_check_ptr: felt}() {

    // Test One
    let x: BigInt3 = BigInt3(100, 99, 98);
    let v: felt = 10;
    let (point) = get_point_from_x(x, v);
    assert point.x.d0 = 100;
    assert point.x.d1 = 99;
    assert point.x.d2 = 98;
    assert point.y.d0 = 50471654703173585387369794;
    assert point.y.d1 = 68898944762041070370364387;
    assert point.y.d2 = 16932612780945290933872774;
    
    return ();
}
