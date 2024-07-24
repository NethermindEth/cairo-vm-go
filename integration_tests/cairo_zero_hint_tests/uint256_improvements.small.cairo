%builtins range_check
from starkware.cairo.common.uint256 import (
    Uint256,
    uint256_check,
    SHIFT,
)

// Adds two integers. Returns the result as a 256-bit integer and the (1-bit) carry.
// Doesn't verify that the result is a valid Uint256
// For use when that check would be performed elsewhere
func _uint256_add_no_uint256_check{range_check_ptr}(a: Uint256, b: Uint256) -> (
    res: Uint256, carry: felt
) {
    alloc_locals;
    local res: Uint256;
    local carry_low: felt;
    local carry_high: felt;
    %{
        sum_low = ids.a.low + ids.b.low
        ids.carry_low = 1 if sum_low >= ids.SHIFT else 0
        sum_high = ids.a.high + ids.b.high + ids.carry_low
        ids.carry_high = 1 if sum_high >= ids.SHIFT else 0
    %}

    assert carry_low * carry_low = carry_low;
    assert carry_high * carry_high = carry_high;

    assert res.low = a.low + b.low - carry_low * SHIFT;
    assert res.high = a.high + b.high + carry_low - carry_high * SHIFT;

    return (res, carry_high);
}

func uint256_sub{range_check_ptr}(a: Uint256, b: Uint256) -> (res: Uint256, sign: felt) {
    alloc_locals;
    local res: Uint256;
    %{
        def split(num: int, num_bits_shift: int = 128, length: int = 2):
            a = []
            for _ in range(length):
                a.append( num & ((1 << num_bits_shift) - 1) )
                num = num >> num_bits_shift
            return tuple(a)

        def pack(z, num_bits_shift: int = 128) -> int:
            limbs = (z.low, z.high)
            return sum(limb << (num_bits_shift * i) for i, limb in enumerate(limbs))

        a = pack(ids.a)
        b = pack(ids.b)
        res = (a - b)%2**256
        res_split = split(res)
        ids.res.low = res_split[0]
        ids.res.high = res_split[1]
    %}
    uint256_check(res);
    let (aa, inv_sign) = _uint256_add_no_uint256_check(res, b);
    assert aa = a;
    return (res, 1 - inv_sign);
}

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
    let x = Uint256(421, 5135);
    let y = Uint256(787, 968);

    // Compute x - y
    let (res, sign) = uint256_sub(x, y);

    assert res = Uint256(340282366920938463463374607431768211090, 4166);
    // x - y >= 0
    assert sign = 1;

    // Compute y - x
    let (res, sign) = uint256_sub(y, x);

    assert res = Uint256(366, 340282366920938463463374607431768207289);
    // y - x < 0
    assert sign = 0;

    let a = 2**64-1;
    let b = 2**64-1;

    uint128_add(a, b);


    return ();
}