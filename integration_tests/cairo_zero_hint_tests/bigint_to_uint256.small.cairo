// The content of this file has been partially borrowed from LambdaClass Cairo VM in Rust

%builtins range_check
from starkware.cairo.common.cairo_secp.bigint3 import BigInt3, SumBigInt3
from starkware.cairo.common.cairo_secp.bigint import nondet_bigint3, bigint_to_uint256

func main{range_check_ptr: felt}() {
    let big_int = BigInt3(d0=9, d1=9, d2=9);

    let (uint256) = bigint_to_uint256(big_int);

    return ();
}
