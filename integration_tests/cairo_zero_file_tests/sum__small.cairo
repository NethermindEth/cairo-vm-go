%builtins output

from starkware.cairo.common.serialize import serialize_word

func sum_four_nums(num1: felt, num2: felt, num3: felt, num4: felt) -> (sum: felt) {
    alloc_locals;
    local sum = num1 + num2 + num3 + num4;
    return (sum=sum);
}

func main{output_ptr: felt*}() {
    alloc_locals;

    const NUM1 = 4;
    const NUM2 = 20;
    const NUM3 = 4;
    const NUM4 = 20;

    let (sum) = sum_four_nums(num1=NUM1, num2=NUM2, num3=NUM3, num4=NUM4);
    serialize_word(sum);
    return ();
}
