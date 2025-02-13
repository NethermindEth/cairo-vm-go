use core::felt252;
use array::ArrayTrait;
use array::SpanTrait;
use core::pedersen::pedersen;
use core::hash::HashStateTrait;
use core::poseidon::{hades_permutation, HashState};

fn main() -> felt252 {
    let mut data: Array<felt252> = ArrayTrait::new();
    data.append(1);
    data.append(2);
    data.append(3);
    data.append(4);
    
    let res = poseidon_hash_span(data.span());
    pedersen(res, 0);

    let mut number = 42;
    number += 100;
    let mut arr = ArrayTrait::new();
    arr.append(1);
    arr.append(2);
    arr.append(3);
    let message_hash: felt252 = 0x503f4bea29baee10b22a7f10bdc82dda071c977c1f25b8f3973d34e6b03b2c;
    let signature_r: felt252 = 0xbe96d72eb4f94078192c2e84d5230cde2a70f4b45c8797e2c907acff5060bb;
    let signature_s: felt252 = 0x677ae6bba6daf00d2631fab14c8acf24be6579f9d9e98f67aa7f2770e57a1f5;
    core::ecdsa::recover_public_key(:message_hash, :signature_r, :signature_s, y_parity: false).unwrap()  
}

// Modified version of poseidon_hash_span that doesn't require builtin gas costs
pub fn poseidon_hash_span(mut span: Span<felt252>) -> felt252 {
    _poseidon_hash_span_inner((0, 0, 0), ref span)
}

/// Helper function for poseidon_hash_span.
fn _poseidon_hash_span_inner(
    state: (felt252, felt252, felt252),
    ref span: Span<felt252>
) -> felt252 {
    let (s0, s1, s2) = state;
    let x = *match span.pop_front() {
        Option::Some(x) => x,
        Option::None => { return HashState { s0, s1, s2, odd: false }.finalize(); },
    };
    let y = *match span.pop_front() {
        Option::Some(y) => y,
        Option::None => { return HashState { s0: s0 + x, s1, s2, odd: true }.finalize(); },
    };
    let next_state = hades_permutation(s0 + x, s1 + y, s2);
    _poseidon_hash_span_inner(next_state, ref span)
}

