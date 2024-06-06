// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust
// See https://github.com/lambdaclass/cairo-vm/

%builtins output range_check
from starkware.cairo.common.serialize import serialize_word

func signed_div_rem{range_check_ptr}(value, div, bound) -> (q: felt, r: felt) {
    let r = [range_check_ptr];
    let biased_q = [range_check_ptr + 1];  // == q + bound.
    let range_check_ptr = range_check_ptr + 2;
    %{
        from starkware.cairo.common.math_utils import as_int, assert_integer
    %}
    let q = biased_q - bound;
    assert value = q * div + r;
    return (q, r);
}

func main{output_ptr: felt*, range_check_ptr: felt}() {
    let (q_negative, r_negative) = signed_div_rem(-10, 3, 29);

    serialize_word(q_negative);
    serialize_word(r_negative);

    let (q, r) = signed_div_rem(10, 3, 29);
    
    return ();
}