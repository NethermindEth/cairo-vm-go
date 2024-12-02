use core::nullable::{nullable_from_box, match_nullable, FromNullableResult};


#[derive(Drop, Copy, Serde)]
struct FP16x16 {
    mag: u32,
    sign: bool
}

fn main() -> Array<felt252> {
    // Create the dictionary
    let mut d: Felt252Dict<Nullable<FP16x16>> = Default::default();

    let box_a = BoxTrait::new(identity(FP16x16 { mag: 1, sign: false }));
    let box_b = BoxTrait::new(identity(FP16x16 { mag: 1, sign: true }));
    let box_c = BoxTrait::new(identity(FP16x16 { mag: 1, sign: true }));

    // Insert it as a `Span`
    d.insert(0, nullable_from_box(box_c));
    d.insert(1, nullable_from_box(box_a));
    d.insert(2, nullable_from_box(box_b));

    // We can't implement Serde for a Felt252Dict due to mutability requirements
    // So we will serialize the dict explicitely
    let mut output: Array<felt252> = ArrayTrait::new();
    // Serialize entry 0
    0.serialize(ref output);
    let array_0 = d.get(0).deref();
    array_0.serialize(ref output);
    // Serialize entry 1
    1.serialize(ref output);
    let array_1 = d.get(1).deref();
    array_1.serialize(ref output);
    // Serialize entry 2
    2.serialize(ref output);
    let array_2 = d.get(2).deref();
    array_2.serialize(ref output);
    // Squash after serializing
    d.squash();
    output
}

// TODO: remove this temporary fix once fixed in cairo
#[inline(never)]
fn identity<T>(t: T) -> T { t }
