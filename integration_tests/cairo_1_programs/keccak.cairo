use core::keccak::cairo_keccak;

fn main() {
    let mut arr = ArrayTrait::new();
    arr.append(1);
    arr.append(2);
    arr.append(3);
    let mut _keccak_result = cairo_keccak(ref arr, 1, 2);
}