from starkware.cairo.common.ec_point import EcPoint

func test_random_ec_point(p: EcPoint, m: felt, q: EcPoint) -> (r: EcPoint) {
    alloc_locals;
    local s: EcPoint;
    %{
        from starkware.crypto.signature.signature import ALPHA, BETA, FIELD_PRIME
        from starkware.python.math_utils import random_ec_point
        from starkware.python.utils import to_bytes

        # Define a seed for random_ec_point that's dependent on all the input, so that:
        #   (1) The added point s is deterministic.
        #   (2) It's hard to choose inputs for which the builtin will fail.
        seed = b"".join(map(to_bytes, [ids.p.x, ids.p.y, ids.m, ids.q.x, ids.q.y]))
        ids.s.x, ids.s.y = random_ec_point(FIELD_PRIME, ALPHA, BETA, seed)
    %}
    return (r=s);
}

func main() {
    let p = EcPoint(
        x=0x6a4beaef5a93425b973179cdba0c9d42f30e01a5f1e2db73da0884b8d6756fc,
        y=0x72565ec81bc09ff53fbfad99324a92aa5b39fb58267e395e8abe36290ebf24f,
    );
    let m = 34;
    let q = EcPoint(
        x=0x654fd7e67a123dd13868093b3b7777f1ffef596c2e324f25ceaf9146698482c,
        y=0x4fad269cbf860980e38768fe9cb6b0b9ab03ee3fe84cfde2eccce597c874fd8,
    );
    test_random_ec_point(p, m, q);
    return ();
}
