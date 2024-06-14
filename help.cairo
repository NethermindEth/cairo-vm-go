%builtins output range_check

from starkware.cairo.common.cairo_secp.signature import get_point_from_x
from starkware.cairo.common.cairo_secp.bigint import BigInt3
from starkware.cairo.common.serialize import serialize_word

func main{output_ptr: felt*, range_check_ptr: felt}() {

    // Test One
    let x_1: BigInt3 = BigInt3(100, 99, 98);
    let v_1: felt = 10;
    let (point_1) = get_point_from_x(x_1, v_1);
    assert point_1.x.d0 = 100;
    assert point_1.x.d1 = 99;
    assert point_1.x.d2 = 98;
    serialize_word(point_1.y.d0);
    serialize_word(point_1.y.d1);
    serialize_word(point_1.y.d2);
    // assert point_1.y.d0 = 50471654703173585387369794;
    // assert point_1.y.d1 = 68898944762041070370364387;
    // assert point_1.y.d2 = 16932612780945290933872774;
    
    return ();
}
