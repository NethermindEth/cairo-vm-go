%builtins output poseidon range_check
from starkware.cairo.common.cairo_builtins import PoseidonBuiltin
from starkware.cairo.common.poseidon_state import PoseidonBuiltinState

func main{output_ptr: felt*, poseidon_ptr: PoseidonBuiltin*, range_check_ptr}() {
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

