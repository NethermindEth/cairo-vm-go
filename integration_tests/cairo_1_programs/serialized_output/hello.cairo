use core::serde::Serde;
use core::option::OptionTrait;

#[derive(Drop, Serde)]
enum MyEnum {
    A: (),
    B: felt252,
}

fn main() -> Array<felt252> {
    let res = MyEnum::B(1234);
    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
