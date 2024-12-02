use core::array::ArrayTrait;


fn main() -> Array<felt252> {
    let mut numbers = ArrayTrait::new();
    numbers.append(1);

    let res = (numbers, 1);
    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
