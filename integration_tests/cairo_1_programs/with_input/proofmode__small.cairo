fn arr_sum(mut vals: Array<felt252>) -> felt252 {
    let mut sum: felt252 = 0;

    loop {
        match vals.pop_front() {
            Option::Some(v) => {
                sum += v;
            },
            Option::None(_) => {
                break sum;
            }
        };
    }
}

fn main(vals: Array<felt252>) -> Array<felt252> {
    let sum1 = arr_sum(vals);
    let mut res: Array<felt252> = ArrayTrait::new();
    res.append(sum1);
    res.append(sum1);
    res
}