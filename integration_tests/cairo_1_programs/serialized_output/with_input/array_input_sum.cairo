use array::ArrayTrait;

fn main(input: Array<felt252>) -> Array<felt252> {
    let mut input = input.span();
    let (index_a, array_a, index_b, array_b): (u32, Array<u32>, u32, Array<u32>) = Serde::deserialize(ref input).unwrap();

    let res = *array_a.at(index_a) + *array_b.at(index_b);

    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
