fn main(input: Array<felt252>) -> Array<felt252> {
    let mut input = input.span();
    let argc: u32 = Serde::deserialize(ref input).unwrap();
    let res = if argc == 0 {
        1_u8
    } else {
        0_u8
    };

    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
