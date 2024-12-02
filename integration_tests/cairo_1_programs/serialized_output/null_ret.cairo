fn main() -> Array<felt252> {
   let _res: Nullable<u32> = null();
   let mut output: Array<felt252> = ArrayTrait::new();
   // Nullable doesn't implement Serde
   ().serialize(ref output);
   output 
}
