use array::ArrayTrait;
use array::SpanTrait;
use pedersen::PedersenTrait;
use core::num::traits::Bounded;

fn check_and_sum(mut vals: Array<felt252>) -> felt252 {
    let mut sum: felt252 = 0;
    let upper_bound: u128 = 1000000; // Example bound
    
    loop {
        match vals.pop_front() {
            Option::Some(v) => {
                // Convert felt252 to u128 for range check
                let v_u128: u128 = v.try_into().unwrap();
                // This will ensure v is non-negative and within u128 bounds
                assert(v_u128 <= upper_bound, 'Value exceeds bound');
                
                sum += v;
            },
            Option::None(_) => {
                break sum;
            }
        };
    }
}

fn main(mut vals: Array<felt252>) -> Array<felt252> {
    // Calculate sum with range checks
    let sum = check_and_sum(vals);
    
    // Hash the sum using Pedersen
    let hash = pedersen::pedersen(sum, 0);
    
    // Create result array
    let mut res: Array<felt252> = ArrayTrait::new();
    
    // Add original sum
    res.append(sum);
    // Add hash of sum
    res.append(hash);
    
    // Add a u128 conversion to demonstrate more range checks
    let sum_u128: u128 = sum.try_into().unwrap();
    res.append(sum_u128.into());
    
    res
}