// inspired from the blake integration tests in the lambdaclass cairo-vm codebase

%builtins range_check bitwise

from starkware.cairo.common.bool import TRUE, FALSE
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.cairo_blake2s.blake2s import blake2s_felts, blake2s, finalize_blake2s
from starkware.cairo.common.cairo_builtins import BitwiseBuiltin
from starkware.cairo.common.uint256 import Uint256

func test_blake2s_felts{range_check_ptr, bitwise_ptr: BitwiseBuiltin*}() {
    alloc_locals;
    let inputs: felt* = alloc();
    assert inputs[0] = 3456722;
    assert inputs[1] = 435425528;
    assert inputs[2] = 3232553;
    assert inputs[3] = 2576195;
    assert inputs[4] = 73471943;
    assert inputs[5] = 17549868;
    assert inputs[6] = 87158958;
    assert inputs[7] = 6353668;
    assert inputs[8] = 343656565;
    assert inputs[9] = 1255962;
    assert inputs[10] = 25439785;
    assert inputs[11] = 1154578;
    assert inputs[12] = 585849303;
    assert inputs[13] = 763502;
    assert inputs[14] = 43753647;
    assert inputs[15] = 74256930;
    let (local blake2s_ptr_start) = alloc();
    let blake2s_ptr = blake2s_ptr_start;

    // Bigendian
    let (result) = blake2s_felts{range_check_ptr=range_check_ptr, blake2s_ptr=blake2s_ptr}(
        16, inputs, TRUE
    );
    assert result.low = 23022179997536219430502258022509199703;
    assert result.high = 136831746058902715979837770794974289597;

    // Little endian
    let (result) = blake2s_felts{range_check_ptr=range_check_ptr, blake2s_ptr=blake2s_ptr}(
        16, inputs, FALSE
    );
    assert result.low = 315510691254085211243916597439546947220;
    assert result.high = 42237338665522721102428636006748876126;
    return ();
}

func test_hash{range_check_ptr, bitwise_ptr: BitwiseBuiltin*}() {
    alloc_locals;
    let inputs: felt* = alloc();
    assert inputs[0] = 'Hell';
    assert inputs[1] = 'o Wo';
    assert inputs[2] = 'rld';
    let (local blake2s_ptr_start) = alloc();
    let blake2s_ptr = blake2s_ptr_start;
    let (output) = blake2s{range_check_ptr=range_check_ptr, blake2s_ptr=blake2s_ptr}(inputs, 9);
    assert output.low = 219917655069954262743903159041439073909;
    assert output.high = 296157033687865319468534978667166017272;
    return ();
}

func fill_array(array: felt*, base: felt, step: felt, array_length: felt, iterator: felt) {
    if (iterator == array_length) {
        return ();
    }
    assert array[iterator] = base + step * iterator;
    return fill_array(array, base, step, array_length, iterator + 1);
}

func test_integration{range_check_ptr, bitwise_ptr: BitwiseBuiltin*}(iter: felt, last: felt) {
    alloc_locals;
    if (iter == last) {
        return ();
    }

    let (data: felt*) = alloc();
    fill_array(data, iter, 2 * iter, 10, 0);

    let (local blake2s_ptr_start) = alloc();
    let blake2s_ptr = blake2s_ptr_start;
    let (res_1: Uint256) = blake2s{range_check_ptr=range_check_ptr, blake2s_ptr=blake2s_ptr}(
        data, 9
    );

    finalize_blake2s(blake2s_ptr_start, blake2s_ptr);

    let (local blake2s_ptr_start) = alloc();
    let blake2s_ptr = blake2s_ptr_start;

    let (data_2: felt*) = alloc();
    assert data_2[0] = res_1.low;
    assert data_2[1] = res_1.high;

    let (res_2) = blake2s_felts{range_check_ptr=range_check_ptr, blake2s_ptr=blake2s_ptr}(
        2, data_2, TRUE
    );

    finalize_blake2s(blake2s_ptr_start, blake2s_ptr);

    if (iter == last - 1 and last == 10) {
        assert res_1.low = 327684140823325841083166505949840946643;
        assert res_1.high = 28077572547397067729112288485703133566;
        assert res_2.low = 323710308182296053867309835081443411626;
        assert res_2.high = 159988406782415793602959692147600111481;
    }

    if (iter == last - 1 and last == 100) {
        assert res_1.low = 26473789897582596397897414631405692327;
        assert res_1.high = 35314462001555260569814614879256292984;
        assert res_2.low = 256911263205530722270005922452382996929;
        assert res_2.high = 248798531786594770765531047659644061441;
    }

    return test_integration(iter + 1, last);
}

func main{range_check_ptr, bitwise_ptr: BitwiseBuiltin*}() {
    test_blake2s_felts();
    test_hash();
    test_integration(0, 10);

    return ();
}
