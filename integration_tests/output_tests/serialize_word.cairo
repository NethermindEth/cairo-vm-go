%builtins output

from starkware.cairo.common.serialize import serialize_word

func main{output_ptr: felt*}() {
    serialize_word('hello');
    serialize_word(540);
    return ();
}
