fn factorial(mut n: felt252) -> felt252 {
    let mut value = n;
    loop {
        if n == 0 || n == 1 || n == 2 {
            break value;
        }

        n -= 1;
        value *= n;
    }
}

fn main() -> Array<felt252> {
    let res = factorial(1000);
    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
