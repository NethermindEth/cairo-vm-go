%builtins output poseidon range_check
from starkware.cairo.common.cairo_builtins import PoseidonBuiltin
from starkware.cairo.common.poseidon_state import PoseidonBuiltinState

func test_poseidon_builtin_random_small_values{output_ptr: felt*, poseidon_ptr: PoseidonBuiltin*}() {
    assert poseidon_ptr[0].input = PoseidonBuiltinState(1, 2, 3);
    let result = poseidon_ptr[0].output;
    let poseidon_ptr = poseidon_ptr + PoseidonBuiltin.SIZE;
    assert result.s0 = 442682200349489646213731521593476982257703159825582578145778919623645026501;
    assert [output_ptr] = result.s0;
    let output_ptr = output_ptr + 1;
    assert result.s1 = 2233832504250924383748553933071188903279928981104663696710686541536735838182;
    assert [output_ptr] = result.s1;
    let output_ptr = output_ptr + 1;
    assert result.s2 = 2512222140811166287287541003826449032093371832913959128171347018667852712082;
    assert [output_ptr] = result.s2;
    let output_ptr = output_ptr + 1;
    return ();
}

func test_poseidon_builtin_random_big_values{output_ptr: felt*, poseidon_ptr: PoseidonBuiltin*}() {
    assert poseidon_ptr[0].input = PoseidonBuiltinState(
        442682200349489646213731521593476982257703159825582578145778919623645026501,
        2233832504250924383748553933071188903279928981104663696710686541536735838182,
        2512222140811166287287541003826449032093371832913959128171347018667852712082
    );
    let result = poseidon_ptr[0].output;
    let poseidon_ptr = poseidon_ptr + PoseidonBuiltin.SIZE;
    assert result.s0 = 3016509350703874362933565866148509373957094754875411937434637891208784994231;
    assert [output_ptr] = result.s0;
    let output_ptr = output_ptr + 1;
    assert result.s1 = 3015199725895936530535660185611704199044060139852899280809302949374221328865;
    assert [output_ptr] = result.s1;
    let output_ptr = output_ptr + 1;
    assert result.s2 = 3062378460350040063467318871602229987911299744598148928378797834245039883769;
    assert [output_ptr] = result.s2;
    let output_ptr = output_ptr + 1;
    return ();
}

func main{output_ptr: felt*, poseidon_ptr: PoseidonBuiltin*, range_check_ptr}() {
    test_poseidon_builtin_random_small_values();
    test_poseidon_builtin_random_big_values();
    return();
}

