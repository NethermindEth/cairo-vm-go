#[derive(Destruct)]
struct NullableVec {
    items: Felt252Dict<Nullable<Box<u32>>>,
    len: usize,
}

fn main() -> Array<felt252> {
    let mut d: Felt252Dict<Nullable<Box<u32>>> = Default::default();

    // Populate the dictionary
    d.insert(0, nullable_from_box(BoxTrait::new(BoxTrait::new(identity(10)))));
    d.insert(1, nullable_from_box(BoxTrait::new(BoxTrait::new(identity(20)))));
    d.insert(2, nullable_from_box(BoxTrait::new(BoxTrait::new(identity(30)))));

    // Return NullableVec
    let mut res = NullableVec {
        items: d,
        len: 3,
    };

   let mut output: Array<felt252> = ArrayTrait::new();
   // Custom Serialization
   // Serialize items field
   // Serialize entry 0
   0.serialize(ref output);
   let val0 = res.items.get(0).deref().unbox();
   val0.serialize(ref output);
   // Serialize entry 1
   1.serialize(ref output);
   let val1 = res.items.get(1).deref().unbox();
   val1.serialize(ref output);
   // Serialize entry 2
   2.serialize(ref output);
   let val2 = res.items.get(2).deref().unbox();
   val2.serialize(ref output);
   // Serialize len field
   res.len.serialize(ref output);
   output 
}

// TODO: remove this temporary fix once fixed in cairo
#[inline(never)]
fn identity<T>(t: T) -> T { t }
