#[derive(Copy, Drop)]
struct Tensor {
    shape: Span<u32>,
    data: Span<u32>
}

fn main(tensor: Tensor) -> u32 {
    *tensor.data.at(0)
}
