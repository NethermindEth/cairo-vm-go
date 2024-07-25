from starkware.cairo.common.uint256 import Uint256

func split_xx(xx: Uint256) -> Uint256{
    alloc_locals;
    local x: Uint256;
    %{
        PRIME = 2**255 - 19
        II = pow(2, (PRIME - 1) // 4, PRIME)

        xx = ids.xx.low + (ids.xx.high<<128)
        x = pow(xx, (PRIME + 3) // 8, PRIME)
        if (x * x - xx) % PRIME != 0:
            x = (x * II) % PRIME
        if x % 2 != 0:
            x = PRIME - x
        ids.x.low = x & ((1<<128)-1)
        ids.x.high = x >> 128
    %}
    return x;
}

func main() {
    let xx: Uint256 = Uint256(7, 17);
    let x = split_xx(xx);

    assert x.low = 316161011683971866381321160306766491472;
    assert x.high = 30265492890921847871084892076606437231;

    let bb: Uint256 = Uint256(1, 1);
    let b = split_xx(bb);

    assert b.low = 60511716334934151406684885798996722026;
    assert b.high = 47999103702266454150683157633393234489;

    let cc: Uint256 = Uint256(28745925095834509, 901854975237520498);
    let c = split_xx(cc);

    assert c.low = 87339391581798877161226027726734745818;
    assert c.high = 135103156151285693363737708431423037815;

    let dd: Uint256 = Uint256(100, 100);
    let d = split_xx(dd);

    assert d.low = 264834796428403050603474250558199008842;
    assert d.high = 139708670101726078043456968902164133435;

    return ();
}