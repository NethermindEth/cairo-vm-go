use array::ArrayTrait;

fn main() -> Array<felt252> {
    let mut numbers = ArrayTrait::new();
    numbers.append(4_u32);
    numbers.append(3_u32);
    numbers.append(2_u32);
    numbers.append(1_u32);
    let res = *numbers.at(1);
    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
