use core::nullable::{nullable_from_box, match_nullable, FromNullableResult};
use core::dict::Felt252DictEntry;

fn main() -> Array<felt252> {
    // Create the dictionary
    let mut d: Felt252Dict<Nullable<Span<felt252>>> = Default::default();

    // Create the array to insert
    let a = array![8, 9, 10, 11];
    let b = array![1, 2, 3];
    let c = array![4, 5, 6];

    // Insert it as a `Span`
    d.insert(66675, nullable_from_box(BoxTrait::new(a.span())));
    d.insert(66676, nullable_from_box(BoxTrait::new(b.span())));
    d.insert(66675, nullable_from_box(BoxTrait::new(c.span())));

    // We can't implement Serde for a Felt252Dict due to mutability requirements
    // So we will serialize the dict explicitely
    let mut output: Array<felt252> = ArrayTrait::new();
    // Serialize entry A
    let key_a = 66675;
    key_a.serialize(ref output);
    let array_a = d.get(key_a).deref();
    array_a.serialize(ref output);
    // Serialize entry B
    let key_b = 66676;
    key_b.serialize(ref output);
    let array_b = d.get(key_b).deref();
    array_b.serialize(ref output);
    // Squash after serializing
    d.squash();
    output
}
