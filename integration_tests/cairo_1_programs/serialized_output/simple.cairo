fn main() -> Array<felt252> {
    let res = true;
    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
