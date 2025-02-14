use core::debug::PrintTrait;

fn main() -> Array<felt252> {
    'Hello, world!'.print();
    1234.print();

    let mut output: Array<felt252> = ArrayTrait::new();
    ().serialize(ref output);
    output
}
