// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/blob/db7fff18c9c1312024ebf4119fcdfeba23657b12/cairo_programs/ed25519_field.cairo

%builtins range_check

from starkware.cairo.common.cairo_secp.bigint import BASE, BigInt3, UnreducedBigInt3, nondet_bigint3
from starkware.cairo.common.cairo_secp.constants import SECP_REM

func verify_zero{range_check_ptr}(val: UnreducedBigInt3) {
    let x = val;
    %{
        from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack

        value = pack(ids.x, PRIME) % SECP_P
    %}
    nondet_bigint3();

    let q = [ap];
    %{
        from starkware.cairo.common.cairo_secp.secp_utils import SECP_P
        q, r = divmod(pack(ids.val, PRIME), SECP_P)
        assert r == 0, f"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}."
        ids.q = q % PRIME
    %}
    let q_biased = [ap + 1];
    q_biased = q + 2 ** 127, ap++;
    [range_check_ptr] = q_biased, ap++;

    tempvar r1 = (val.d0 + q * SECP_REM) / BASE;
    assert [range_check_ptr + 1] = r1 + 2 ** 127;

    tempvar r2 = (val.d1 + r1) / BASE;
    assert [range_check_ptr + 2] = r2 + 2 ** 127;

    assert val.d2 = q * (BASE / 4) - r2;

    let range_check_ptr = range_check_ptr + 3;
    return ();
}

func verify_zero_alt{range_check_ptr}(val: UnreducedBigInt3) {
    let x = val;
    %{
        from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack

        value = pack(ids.x, PRIME) % SECP_P
    %}
    nondet_bigint3();

    let q = [ap];
    %{
        from starkware.cairo.common.cairo_secp.secp_utils import pack

        q, r = divmod(pack(ids.val, PRIME), SECP_P)
        assert r == 0, f"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}."
        ids.q = q % PRIME
    %}
    let q_biased = [ap + 1];
    q_biased = q + 2 ** 127, ap++;
    [range_check_ptr] = q_biased, ap++;

    tempvar r1 = (val.d0 + q * SECP_REM) / BASE;
    assert [range_check_ptr + 1] = r1 + 2 ** 127;

    tempvar r2 = (val.d1 + r1) / BASE;
    assert [range_check_ptr + 2] = r2 + 2 ** 127;

    assert val.d2 = q * (BASE / 4) - r2;

    let range_check_ptr = range_check_ptr + 3;
    return ();
}

func test_verify_zero{range_check_ptr: felt}() {
    let val = UnreducedBigInt3(0, 0, 0);

    verify_zero(val);

    return ();
}

func test_verify_zero_alt{range_check_ptr: felt}() {
    let val = UnreducedBigInt3(0, 0, 0);

    verify_zero_alt(val);

    return ();
}

func main{range_check_ptr: felt}() {
    test_verify_zero();
    test_verify_zero_alt();

    return ();
}
