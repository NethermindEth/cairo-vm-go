// This file has been borrowed from https://github.com/dojoengine/cairo-rs/blob/bb491f2a9ea0514bbeba92d858b28baaf41053e7/cairo_programs/simple_print.cairo#L4

%builtins output

from starkware.cairo.common.serialize import serialize_word

func main{output_ptr: felt*}() {
    let x = 100;

    let y = x / 2;

    serialize_word(y);

    ret;
}
