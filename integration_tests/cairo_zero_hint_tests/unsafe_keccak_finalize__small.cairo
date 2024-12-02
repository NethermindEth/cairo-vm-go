%builtins output

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.keccak import unsafe_keccak_finalize, KeccakState
from starkware.cairo.common.uint256 import Uint256

func main{output_ptr: felt*}() {
    alloc_locals;

    let (data: felt*) = alloc();

    assert data[0] = 0;
    assert data[1] = 1;
    assert data[2] = 2;

    let keccak_state = KeccakState(start_ptr=data, end_ptr=data + 2);

    let res: Uint256 = unsafe_keccak_finalize(keccak_state);

    return ();
}