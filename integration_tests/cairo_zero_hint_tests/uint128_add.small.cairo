%builtins range_check

// Source: https://github.com/NethermindEth/research-basic-Cairo-operations-big-integers/blob/fe1ddf69549354a4f241074486db4cd9fb259d51/lib/uint256_improvements.cairo


from starkware.cairo.common.uint256 import (
    Uint256,
    SHIFT,
    HALF_SHIFT,
    split_64,
    uint256_check,
    uint256_add,
    uint256_le,
    uint256_lt,
)


// assumes inputs are <2**128
func uint128_add{range_check_ptr}(a: felt, b: felt) -> (result: Uint256) {
    alloc_locals;
    local carry: felt;
    %{
        res = ids.a + ids.b
        ids.carry = 1 if res >= ids.SHIFT else 0
    %}
    // Either 0 or 1
    assert carry * carry = carry;
    local res = a + b - carry * SHIFT;
    [range_check_ptr] = res;
    let range_check_ptr = range_check_ptr + 1;

    return (result=Uint256(low=res, high=carry));
}


func main{range_check_ptr}() {

    let a = 2**64-1;
    let b = 2**64-1;

    uint128_add(a, b);

    return ();
}