%builtins range_check
from starkware.cairo.common.usort import usort
from starkware.cairo.common.alloc import alloc

func main{range_check_ptr}() -> () {
    alloc_locals;
    let (input_array: felt*) = alloc();
    assert input_array[0] = 8;
    assert input_array[1] = 9;
    assert input_array[2] = 7;

    let (output_len, output, multiplicities) = usort(input_len=3, input=input_array);

    assert output_len = 3;
    assert output[0] = 7;
    assert output[1] = 8;
    assert output[2] = 9;
    assert multiplicities[0] = 1;
    assert multiplicities[1] = 1;
    assert multiplicities[2] = 1;

    let (input_array: felt*) = alloc();
    assert input_array[0] = 11;
    assert input_array[1] = 24;
    assert input_array[2] = 99;
    assert input_array[3] = 2;
    assert input_array[4] = 66;
    assert input_array[5] = 49;
    assert input_array[6] = 11;
    assert input_array[7] = 23;
    assert input_array[8] = 88;
    assert input_array[9] = 7;

    let (output_len, output, multiplicities) = usort(input_len=10, input=input_array);

    assert output_len = 9;
    assert output[0] = 2;
    assert output[1] = 7;
    assert output[2] = 11;
    assert output[3] = 23;
    assert output[4] = 24;
    assert output[5] = 49;
    assert output[6] = 66;
    assert output[7] = 88;
    assert output[8] = 99;

    assert multiplicities[0] = 1;
    assert multiplicities[1] = 1;
    assert multiplicities[2] = 2;
    assert multiplicities[3] = 1;
    assert multiplicities[4] = 1;
    assert multiplicities[5] = 1;
    assert multiplicities[6] = 1;
    assert multiplicities[7] = 1;
    assert multiplicities[8] = 1;
    
    return ();
}
