%builtins range_check

from starkware.cairo.common.cairo_secp.bigint import BigInt3
from starkware.cairo.common.cairo_secp.ec import EcPoint, ec_negate, compute_doubling_slope, compute_slope, ec_double, fast_ec_add

func test_ec_negate{range_check_ptr}() {
    let p = EcPoint(BigInt3(1, 2, 3), BigInt3(1, 2, 3));
    let (res) = ec_negate(p);

    let p = EcPoint(
        BigInt3(12424, 53151, 363737),
        BigInt3(77371252455336267181195244, 77371252455336267181195261, 9671406556917033397649404),
    );
    let (res) = ec_negate(p);

    return ();
}

func test_compute_doubling_slope{range_check_ptr}() {
    let p = EcPoint(BigInt3(1, 2, 3), BigInt3(1, 2, 3));
    let (res) = compute_doubling_slope(p);

    let p = EcPoint(
        BigInt3(12424, 53151, 363737),
        BigInt3(77371252455336267181195244, 77371252455336267181195261, 9671406556917033397649404),
    );
    let (res) = compute_doubling_slope(p);

    return ();
}

func test_compute_slope{range_check_ptr}() {
    let p1 = EcPoint(BigInt3(1, 2, 3), BigInt3(1, 2, 3));
    let p2 = EcPoint(
        BigInt3(12424, 53151, 363737),
        BigInt3(77371252455336267181195244, 77371252455336267181195261, 9671406556917033397649404),
    );
    let (res) = compute_slope(p1, p2);

    let p1 = EcPoint(
        BigInt3(17117865558768631194064792, 12501176021340589225372855, 9198697782662356105779718),
        BigInt3(6441780312434748884571320, 57953919405111227542741658, 5457536640262350763842127)
    );
    let p2 = EcPoint(
        BigInt3(12424, 53151, 363737),
        BigInt3(77371252458936267181195765, 77561252455336267181195987, 9674506556917033397649657),
    );
    let (res) = compute_slope(p1, p2);

    return ();
}

func test_ec_double{range_check_ptr}() {
    let p = EcPoint(BigInt3(1, 2, 3), BigInt3(1, 2, 3));
    let (res) = ec_double(p);

    let p = EcPoint(
        BigInt3(12424, 53151, 363737),
        BigInt3(77371252455336267181195244, 77371252455336267181195261, 9671406556917033397649404),
    );
    let (res) = ec_double(p);

    return ();
}

func test_fast_ec_add{range_check_ptr}() {
    let p1 = EcPoint(BigInt3(1, 2, 3), BigInt3(1, 2, 3));
    let p2 = EcPoint(
        BigInt3(12424, 53151, 363737),
        BigInt3(77371252455336267181195244, 77371252455336267181195261, 9671406556917033397649404),
    );
    let (res) = fast_ec_add(p1, p2);

    let p1 = EcPoint(
        BigInt3(17117865558768631194064792, 12501176021340589225372855, 9198697782662356105779718),
        BigInt3(6441780312434748884571320, 57953919405111227542741658, 5457536640262350763842127)
    );
    let p2 = EcPoint(
        BigInt3(12424, 53151, 363737),
        BigInt3(77371252458936267181195765, 77561252455336267181195987, 9674506556917033397649657),
    );
    let (res) = fast_ec_add(p1, p2);

    return ();
}

func main{range_check_ptr}() {
    test_ec_negate();
    test_compute_doubling_slope();
    test_compute_slope();
    test_ec_double();
    test_fast_ec_add();
    return ();
}
