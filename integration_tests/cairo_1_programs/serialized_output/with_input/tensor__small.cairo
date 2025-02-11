#[derive(Copy, Drop, Serde)]
struct Tensor {
    shape: Span<u32>,
    data: Span<u32>
}
fn main(input: Array<felt252>) -> Array<felt252> {
    let mut input = input.span();
    let tensor : Tensor = Serde::deserialize(ref input).unwrap();
    let res = *tensor.data.at(0);

    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
