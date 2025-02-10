use core::nullable::{nullable_from_box, match_nullable, FromNullableResult};

fn main() -> Array<felt252>{
    // Create the dictionary
    let mut d: Felt252Dict<Nullable<Span<felt252>>> = Default::default();

    // Create the array to insert
    let a = array![8, 9, 10];

    // Insert it as a `Span`
    d.insert(0, nullable_from_box(BoxTrait::new(a.span())));

    let mut output: Array<felt252> = ArrayTrait::new();
    // Felt252Dict doesn't implement Serde
    ().serialize(ref output);
    output

}
