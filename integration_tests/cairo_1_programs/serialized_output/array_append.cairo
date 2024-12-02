use array::ArrayTrait;

fn main() -> Array<felt252> {
    let mut numbers = ArrayTrait::new();
    numbers.append(4_u32);
    numbers.append(2_u32);
    let _x = numbers.pop_front();
    let mut output: Array<felt252> = ArrayTrait::new();
    numbers.serialize(ref output);
    output
}
