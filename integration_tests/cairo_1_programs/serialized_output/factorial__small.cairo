use core::felt252;

fn main() -> Array<felt252> {
    let n = 10;
    let result = factorial(n);

    let mut output: Array<felt252> = ArrayTrait::new();
    result.serialize(ref output);
    output
}

fn factorial(n: felt252) -> felt252 {
    match n {
        0 => 1,
        _ => n * factorial(n - 1),
    }
}
