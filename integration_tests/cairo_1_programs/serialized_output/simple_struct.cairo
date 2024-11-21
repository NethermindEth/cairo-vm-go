#[derive(Drop, Serde)]
struct Hello {
    a: felt252
}

fn main() -> Array<felt252> {
    let res = Hello {
        a: 100
    };
    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output
}
