%builtins output

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.keccak import unsafe_keccak

func main{output_ptr: felt*}() {
    alloc_locals;

    let (data: felt*) = alloc();

    assert data[0] = 500;
    assert data[1] = 2;
    assert data[2] = 3;
    assert data[3] = 6;
    assert data[4] = 1;
    assert data[5] = 4444;

    let (low: felt, high: felt) = unsafe_keccak(data, 6);

    return ();
}