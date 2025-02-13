use core::felt252;
use array::ArrayTrait;
use core::Serde;


fn main() -> Array<felt252> {
    let mut output: Array<felt252> = ArrayTrait::new();
    let a : u32 = 10 - 2;
    a.serialize(ref output);
    output
}
