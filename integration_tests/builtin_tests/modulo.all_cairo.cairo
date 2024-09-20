%builtins range_check add_mod mul_mod
// TODO: Import directly from common library once released
from cairo_programs.mod_builtin_feature.common.modulo import ModBuiltin, UInt384, run_mod_p_circuit
// from starkware.common.cairo_builtins import ModBuiltin, UInt384
// from starkware.cairo.common.modulo import run_mod_p_circuit
from starkware.cairo.common.registers import get_label_location
from starkware.cairo.common.alloc import alloc

func main{range_check_ptr, add_mod_ptr: ModBuiltin*, mul_mod_ptr: ModBuiltin*}() {
    alloc_locals;

    let p = UInt384(d0=1, d1=1, d2=0, d3=0);
    let x1 = UInt384(d0=1, d1=0, d2=0, d3=0);
    let x2 = UInt384(d0=2, d1=1, d2=0, d3=0);
    let x3 = UInt384(d0=2, d1=0, d2=0, d3=0);
    let res = UInt384(d0=2, d1=0, d2=0, d3=0);

    let (local values_arr: UInt384*) = alloc();
    assert values_arr[0] = x1;
    assert values_arr[1] = x2;
    assert values_arr[2] = x3;
    assert values_arr[7] = res;

    let (local add_mod_offsets_arr: felt*) = alloc();
    assert add_mod_offsets_arr[0] = 0;   // x1
    assert add_mod_offsets_arr[1] = 12;  // x2 - x1
    assert add_mod_offsets_arr[2] = 4;   // x2
    assert add_mod_offsets_arr[3] = 16;  // (x2 - x1) * x3
    assert add_mod_offsets_arr[4] = 20;  // x1 * x3
    assert add_mod_offsets_arr[5] = 24;  // (x2 - x1) * x3 + x1 * x3

    let (local mul_mod_offsets_arr: felt*) = alloc();
    assert mul_mod_offsets_arr[0] = 12;  // x2 - x1
    assert mul_mod_offsets_arr[1] = 8;   // x3
    assert mul_mod_offsets_arr[2] = 16;  // (x2 - x1) * x3
    assert mul_mod_offsets_arr[3] = 0;   // x1
    assert mul_mod_offsets_arr[4] = 8;   // x3
    assert mul_mod_offsets_arr[5] = 20;  // x1 * x3
    assert mul_mod_offsets_arr[6] = 8;   // x3
    assert mul_mod_offsets_arr[7] = 28;  // ((x2 - x1) * x3 + x1 * x3) / x3 = x2 mod p
    assert mul_mod_offsets_arr[8] = 24;  // (x2 - x1) * x3 + x1 * x3

    run_mod_p_circuit(
        p=p,
        values_ptr=values_arr,
        add_mod_offsets_ptr=add_mod_offsets_arr,
        add_mod_n=2,
        mul_mod_offsets_ptr=mul_mod_offsets_arr,
        mul_mod_n=3,
    );
    
    return ();
}
