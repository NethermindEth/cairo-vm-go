// This file has been borrowed from https://github.com/ftupas/encode-cairo-bootcamp/blob/221ad50eb12ce7635702f5e673b0a9a9dbcf1e58/homework_3/homework_3.cairo#L4

%builtins output

from starkware.cairo.common.serialize import serialize_word

func square(x: felt) -> (x_squared: felt) {
    return (x_squared=x * x);
}

func main{output_ptr: felt*}() {
    tempvar x = 10;
    tempvar y = x + x;
    tempvar z = y * y + x;
    serialize_word(x);
    serialize_word(y);
    serialize_word(z);

    let (x_squared: felt) = square(x);
    serialize_word(x_squared);
    return ();
}
