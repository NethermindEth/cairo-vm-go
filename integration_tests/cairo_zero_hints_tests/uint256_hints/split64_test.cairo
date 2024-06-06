// The content of this file has been partially borrowed from cairo-lang implementation in python
// See https://github.com/starkware-libs/cairo-lang

// Splits a field element in the range [0, 2^192) to its low 64-bit and high 128-bit parts.
// Soundness guarantee: a is in the range [0, 2^192).

const HALF_SHIFT = 2 ** 64;

func split_64(a: felt) -> (low: felt, high: felt) {
    alloc_locals;
    local low: felt;
    local high: felt;

    %{
        ids.low = ids.a & ((1<<64) - 1)
        ids.high = ids.a >> 64
    %}
    return (low, high);
}

func main() {
    let (a, b) = split_64(8746);
    assert a = 8746;
    assert b = 0;
    let (c,d) = split_64(2**127);
    assert c = 0;
    assert d = 2 ** 63;
    let (e,f) = split_64(-5);
    assert e = 18446744073709551612;
    assert f = 196159429230833779654668657131193454380566933979560673279;
    return ();
}
