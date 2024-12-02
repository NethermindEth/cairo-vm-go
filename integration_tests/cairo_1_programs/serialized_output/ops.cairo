fn main() -> Array<felt252> {
    let res = 2_u32 + 4_u32;
    let mut output: Array<felt252> = ArrayTrait::new();
   res.serialize(ref output);
   output 
}
