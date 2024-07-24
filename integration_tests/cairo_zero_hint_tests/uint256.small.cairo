%builtins range_check

from starkware.cairo.common.uint256 import (Uint256, uint256_add, split_64, uint256_sqrt, uint256_signed_nn, uint256_unsigned_div_rem, uint256_mul)
from starkware.cairo.common.alloc import alloc

func fill_array{range_check_ptr: felt}(array: Uint256*, base: Uint256, step: Uint256, array_length: felt, iterator: felt):
    if iterator == array_length:
        return()
    end
    let (res, carry_high) = uint256_add(step, base)
    let (sqrt) = uint256_sqrt(res)

    assert array[iterator] = sqrt
    return fill_array(array, base, array[iterator], array_length, iterator + 1)
end

func main{range_check_ptr: felt}():
    let x :Uint256 = Uint256(5,2)
    let y = Uint256(3,7)
    let (res, carry_high) = uint256_add(x,y)
    assert res.low = 8
    assert res.high = 9
    assert carry_high = 0

    let (low, high) = split_64(850981239023189021389081239089023)
    assert low = 7249717543555297151
    assert high = 46131785404667
    
    let (root) = uint256_sqrt(Uint256(17,7))
    assert root = Uint256(48805497317890012913,0)

    let (signed_nn) = uint256_signed_nn(Uint256(5,2))
    assert signed_nn = 1
    let (p) = uint256_signed_nn(Uint256(1,170141183460469231731687303715884105728))
    assert p = 0
    let (q) = uint256_signed_nn(Uint256(1,170141183460469231731687303715884105727))
    assert q = 1

    let (a_quotient, a_remainder) = uint256_unsigned_div_rem(Uint256(89,72), Uint256(3,7))
    assert a_quotient = Uint256(10,0)
    assert a_remainder = Uint256(59,2)

    let (b_quotient, b_remainder) = uint256_unsigned_div_rem(Uint256(-3618502788666131213697322783095070105282824848410658236509717448704103809099,2), Uint256(5,2))
    assert b_quotient = Uint256(1,0)
    assert b_remainder = Uint256(340282366920938463463374607431768211377,0)
    
    let (mult_low_a, mult_high_a) = uint256_mul(Uint256(59,2),  Uint256(10,0))
    assert mult_low_a = Uint256(590,20)
    assert mult_high_a = Uint256(0,0)

    let (mult_low_b: Uint256, mult_high_b: Uint256) = uint256_mul(Uint256(271442546951262198976322048597925888860,0), Uint256(271442546951262198976322048597925888860,0))
    assert mult_low_b = Uint256(42047520920204780886066537579778623760, 216529163594619381764978757921136443390)
    assert mult_high_b = Uint256(0,0)

    let array_length = 100
    let (sum_array : Uint256*) = alloc()
    fill_array(sum_array, Uint256(57,8), Uint256(17,7), array_length, 0)

    return()
end