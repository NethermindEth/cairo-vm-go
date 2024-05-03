---
sidebar_position: 2
---

# Builtins

Regular Cairo instructions are like regular Legos, they work, but for complex things they can be slow. Builtins are special Lego pieces that snap together easily, like pre-built wheels or doors. This makes building things much faster. In a Cairo context, builtins help write clean and efficient Cairo programs, just like these special Lego pieces help build cool things faster.

This section provides an overview of the builtins functions available in the Cairo Virtual Machine (Cairo VM).

### What is builtins in cairo?

Builtins are predefined optimized low-level execution units which are added to the Cairo CPU board to perform predefined computations which are expensive to perform in Cairo.

### Builtins and Cairo memory

Communication between the CPU and built-in functionalities occurs via memory-mapped I/O. Each builtin is allocated a contiguous memory region and enforces specific constraints on the data residing within that area. The Pedersen builtin is a good example to explain this.

Pedersen builtin establishes that:

```
[p + 2] = hash([p + 0], [p + 1])
[p + 5] = hash([p + 3], [p + 4])
[p + 8] = hash([p + 6], [p + 7])
```

The Cairo code may read or write from this memory cells to “invoke” the builtin. The following code verifies that `hash(x, y) == z`

```
// Write the value of x to [p + 0].
x=[p]; // Write the value of y to [p + 1].
y=[p + 1]; // The builtin makes that [p + 2] == hash([p + 0], [p + 1]).
z=[p + 2];
```

Cairo memory immutability requires a careful usage of builtin instances because memory locations like `[p + 0]`, `[p + 1]`, and `[p + 2]` are used for a single hash computation, they cannot be reused for subsequent calculations and needs to tracking an unused memory location pointer (`hash_ptr`).

By convention, the functions that utilizing builtins receive a pointer to the unused memory location as an argument and return an updated pointer reflecting the next available memory slot. For example:

```
func hash2(hash_ptr: felt*, x, y) -> (hash_ptr: felt*, z: felt) {
  // Invoke the hash function.
  x = [hash_ptr];
  y = [hash_ptr + 1];
  // Update pointer (increment by 3) and return result.
  return (hash_ptr=hash_ptr + 3, z=[hash_ptr + 2]);
}
```

On the other hand, Starkware Cairo library provides typed references for improved readability. Here's the code rewritten with `HashBuiltin` type like this example:

```
from starkware.cairo.common.cairo_builtins import HashBuiltin

func hash2(hash_ptr: HashBuiltin*, x, y) -> (hash_ptr: HashBuiltin*, z: felt) {
  let hash = hash_ptr;
  // Invoke the hash function using typed access.
  hash.x = x;
  hash.y = y;
  // Update pointer and return result.
  return (hash_ptr=hash_ptr + HashBuiltin.SIZE, z=hash.result);
}
```

### Some specific builtins in Cairo VM

- **output**
    
    The output builtin, accessed with a pointer to type felt, is used for writing program outputs.
    
    ```rust
    %builtins output 
    //allows the program to write this output to a designated memory location.
    from starkware.cairo.common.serialize import serialize_word
    func main{output_ptr: felt*}() { 
    //the output here indicate that parameter is a pointer to a memory location of 
    //type felt*. The location is used to store the program output 
        serialize_word(1234);
        serialize_word(4321);
        return ();
    }
    ```
    
- **pedersen**
    
    The pedersen builtin, accessed with a pointer to type HashBuiltin, is used for pedersen hashing computations.
    
    ```rust
    // Function to perform Pedersen hash on two numbers
    // This function takes two felt values (a and b) as input.
    func pedersen_hash(a felt, b felt) -> (HashBuiltin) {
    	// Call the `pedersen` builtin to perform the Pedersen hash computation on the two inputs (a and b).
      // The `pedersen` builtin is likely accessed with a pointer to a type named `HashBuiltin` 
      // (specific names might vary depending on Cairo version).
      
      let hash = pedersen(a, b)
      // The result of the Pedersen hash is stored in a variable named `hash`. 
      // The type of `hash` is likely `HashBuiltin`.
      
      // Return the calculated hash value.
      return hash
    }
    ```
    
- **range_check**
    
    Unlike other builtins, the range_check is accessed using a type felt, rather than a pointer. This builtin is mostly used for integer comparisons, and facilitates check to confirm that a field element is within a range `[0, 2^128)`
    
    ```rust
    // Function to check if a number is within the valid range
    func is_valid_range(x felt) -> (felt) {
      // Use the range_check builtin to check the range
      if !range_check(x) {
        // Handle the case where the value is out of range (optional)
      }
      return x
    }
    ```
    
- **ecdsa**
    
    The ecdsa builtin, accessed with a pointer to type SignatureBuiltin, is used for verifying ECDSA signatures on a message using  public key.
    
    ```rust
    // Function to verify an ECDSA signature (complex example)
    func verify_signature(message felt, signature SignatureBuiltin, public_key felt) -> (felt) {
      // This example requires additional logic to unpack the signature and public key
      // and call the ecdsa builtin for verification. Refer to Cairo documentation for details.
      let is_valid = ecdsa(message, signature, public_key)
      // Handle the result of the verification (valid or invalid)
      return is_valid
    }
    ```
    
- **bitwise**
    
    The bitwise builtin, accessed with a pointer to type BitwiseBuiltin, is used for carrying out bitwise operations on felts.
    
    ```rust
    // Function to perform a bitwise AND operation on two numbers
    func bitwise_and(a felt, b felt) -> (felt) {
    //it checks if each bit position in a and b is a 1 at the same time  
      let result = bitwise(BitwiseBuiltin.AND, a, b)
      return result
    }
    ```