fn main() -> Array<felt252> {
    let a = 1234_u128;
    let b = 5678_u128;

    let c0 = a & b;
    let c1 = a ^ b;
    let c2 = a | b;

    let c3 = c0 + c1 + c2;
    let mut output: Array<felt252> = ArrayTrait::new();
    c3.serialize(ref output);
    output
}
