%builtins output

func main{output_ptr}() {
    assert [output_ptr] = 9;

    // Manually update the output builtin pointer.
    let output_ptr = output_ptr + 1;

    // output_ptr will be implicitly returned.
    return ();
}
