fn main() -> Array<felt252> {
    let a: u128 = 123;
    let b: bytes31 = a.into();
    let mut output: Array<felt252> = ArrayTrait::new();
    b.serialize(ref output);
    output
}
