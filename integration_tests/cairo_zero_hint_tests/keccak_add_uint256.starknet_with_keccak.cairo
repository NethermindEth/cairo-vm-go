%builtins output range_check bitwise

from starkware.cairo.common.keccak_utils.keccak_utils import keccak_add_uint256
from starkware.cairo.common.uint256 import Uint256
from starkware.cairo.common.cairo_builtins import BitwiseBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.serialize import serialize_word

func main{output_ptr: felt*, range_check_ptr, bitwise_ptr: BitwiseBuiltin*}() {
    alloc_locals;

    let (inputs) = alloc();
    let inputs_start = inputs;

    let num = Uint256(34623634663146736, 598249824422424658356);

    keccak_add_uint256{inputs=inputs_start}(num=num, bigend=0);

    return ();
}