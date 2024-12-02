use array::ArrayTrait;

fn main(input: Array<felt252>) -> Array<felt252> {
    let mut input = input.span();
    let (array_a, array_b): (Array<u32>, Array<u32>) = Serde::deserialize(ref input).unwrap();

    let res = array_a.len() + array_b.len();
    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
