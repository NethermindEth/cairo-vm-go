%builtins range_check

func main(range_check_ptr: felt) -> (range_check_ptr: felt) {
    assert [range_check_ptr] = -1;
    return (range_check_ptr=range_check_ptr + 1);
}
