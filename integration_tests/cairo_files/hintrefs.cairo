// This test uses the artificially constructed TestAssignCode hint
// to test different hint refs evaluation.
//
// Even if the hint's code is the same in every function below,
// the referenced ids.a always has an address that requires
// a different way of computation (see per-func comments).
// They also usually have different ApTracking values associated with them.
//
// See https://github.com/NethermindEth/cairo-vm-go/issues/197

// [cast(fp, felt*)]
func simple_fp_ref() -> felt {
    alloc_locals;
    local a = 43;
    %{ memory[ap] = ids.a %}
    return [ap];
}

// [cast(ap + (-1), felt*)]
func ap_with_offset() -> felt {
    [ap] = 0, ap++;
    [ap] = 10, ap++;
    tempvar a = 32;
    [ap] = 100, ap++;
    [ap] = 200, ap++;
    %{ memory[ap] = ids.a %}
    return [ap];
}

// cast([fp + (-4)] + [fp + (-3)], felt)
func fp_args_sum(arg1: felt, arg2: felt) -> felt {
    let a = arg1 + arg2;
    %{ memory[ap] = ids.a %}
    return [ap];
}

// cast([ap + (-1)] + [fp + 1], felt)
func ap_plus_fp_deref() -> felt {
    alloc_locals;
    local l1 = 11;
    local l2 = 22; // [fp+1]
    local l3 = 33;
    tempvar t1 = 111;
    tempvar t2 = 222;
    tempvar t3 = 333; // [ap-1]
    let a = [ap-1] + [fp+1];
    %{ memory[ap] = ids.a %}
    return [ap]; // 355
}

func main() {
    alloc_locals;

    local v1 = simple_fp_ref();
    [ap] = v1, ap++;
    local v2 = ap_with_offset();
    [ap] = v2, ap++;
    local v3 = fp_args_sum(4, 6);
    [ap] = v3, ap++;
    local v4 = ap_plus_fp_deref();
    [ap] = v4, ap++;

    ret;
}
