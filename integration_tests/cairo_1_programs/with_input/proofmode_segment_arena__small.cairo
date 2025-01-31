use array::ArrayTrait;
use array::SpanTrait;
use dict::Felt252DictTrait;
use pedersen::PedersenTrait;

fn main(mut vals: Array<felt252>) -> Array<felt252> {
    // Create a dictionary using segment arena
    let mut dict = felt252_dict_new::<felt252>();
    
    // Store each value in the dictionary with its index
    let mut index: felt252 = 0;
    loop {
        match vals.pop_front() {
            Option::Some(v) => {
                dict.insert(index, v);
                index += 1;
            },
            Option::None(_) => {
                break;
            }
        };
    };
    
    // Create result array
    let mut res: Array<felt252> = ArrayTrait::new();
    
    // Read values back from dictionary and hash them
    let mut i: felt252 = 0;
    loop {
        let i_u128: u128 = i.try_into().unwrap();
        let index_u128: u128 = index.try_into().unwrap();

        if i_u128 >= index_u128 {
            break;
        }
        
        let value = dict.get(i);
        let hash = pedersen::pedersen(value, i);
        res.append(hash);
        
        i += 1;
    };
    
    res
}