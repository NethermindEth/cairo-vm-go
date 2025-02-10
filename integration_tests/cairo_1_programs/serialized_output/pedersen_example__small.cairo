use core::pedersen::pedersen;

fn main() -> Array<felt252> {
    let res = pedersen(1, 0);
    let mut output: Array<felt252> = ArrayTrait::new();
    res.serialize(ref output);
    output 
}
