---
sidebar_position: 2
---
# Builtins Documentation

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
    // %builtins directive specifies which builtins are used by the program
    %builtins output pedersen 
    
    //Import necessary built-in functions for cryptographic hashing.
    from starkware.cairo.common.cairo_builtins import HashBuiltin
    from starkware.cairo.common.hash import hash2
    
    //It takes two pointers: output_ptr and pedersen_ptr.
    //output_ptr points to the output built-in function.
    //pedersen_ptr points to the pedersen built-in function.
    func main{output_ptr, pedersen_ptr: HashBuiltin*}() {
        // This line implicitly updates the pedersen_ptr reference to pedersen_ptr + 3.
        let (res) = hash2{hash_ptr=pedersen_ptr}(1, 2);
        assert [output_ptr] = res;
    
        //Update the output builtin pointer.
        let output_ptr = output_ptr + 1;
    
        // output_ptr and pedersen_ptr will be implicitly returned.
        return ();
    }
    ```
    
- **range_check**
    
    The range-check builtin is utilized to ensure that a field element falls within the specified range **`[0, 2^128)`**. This check is performed by examining three elements located at consecutive addresses starting from **`p`**, ensuring each falls within the range.
    
    **`0 <= [p + 0] < 2^128`**
    
    **`0 <= [p + 1] < 2^128`**
    
    **`0 <= [p + 2] < 2^128`**
    
    Here, **`p`** represents the starting address of the built-in.
    
    To check if a value **`x`** falls within a smaller range **`[0,BOUND]`** where **`BOUND`** is less than **`2^128`**, two instances of the range-check are used:
    
    1. One instance confirms **`0 ≤ x < 2^128`**.
    2. Another instance validates **`0 ≤ BOUND - x < 2^128`**.
    
    This approach facilitates computing integer division with a remainder using the range-check built-in. The objective is to calculate **`q = [x/y]`** and **`r = x mod y`**, rewritten as **`x = q*y + r`**, ensuring **`0 ≤ r < y`** holds true.
    Care must be taken to prevent overflow during the **`x = q*y + r`** calculation. For simplicity, it's assumed that **`0 ≤ x, y < 2^64`**. Adjustments can be made to the code based on specific constraints if this assumption doesn't hold true.
    
    The provided code computes **`q`** and **`r`**, while ensuring **`0 ≤ x, y < 2^64`**, provided **`|F| > 2^128`**.
    
    ```rust
    %builtins range_check
    func div{range_check_ptr}(x, y) -> (q: felt, r: felt) {
        alloc_locals;
        local q;
        local r;
        %{ ids.q, ids.r = ids.x // ids.y, ids.x % ids.y %}
    
        // Check that 0 <= x < 2**64.
        [range_check_ptr] = x;
        assert [range_check_ptr + 1] = 2 ** 64 - 1 - x;
    
        // Check that 0 <= y < 2**64.
        [range_check_ptr + 2] = y;
        assert [range_check_ptr + 3] = 2 ** 64 - 1 - y;
    
        // Check that 0 <= q < 2**64.
        [range_check_ptr + 4] = q;
        assert [range_check_ptr + 5] = 2 ** 64 - 1 - q;
    
        // Check that 0 <= r < y.
        [range_check_ptr + 6] = r;
        assert [range_check_ptr + 7] = y - 1 - r;
    
        // Verify that x = q * y + r.
        assert x = q * y + r;
    
        let range_check_ptr = range_check_ptr + 8;
        return (q=q, r=r);
    }
    ```
    
- **ecdsa**
    
    A structure defines the memory layout for the signature builtin. It's employed by various functions in the common library, including the ecdsa builtin. For instance, in the verify_ecdsa_signature() function, there's an implicit argument of type SignatureBuiltin, which points to a variable of type SignatureBuiltin
    
    The struct comprises two members of type felt:
    
    - **pub_key**: Represents an ECDSA public key.
    - **message**: Denotes a message signed by the pub_key.
    
    Additionally, there's a pointer named ecdsa_ptr, which is of type SignatureBuiltin. This pointer is used to manage the signature builtin instances internally.
    
    ```rust
    from starkware.cairo.common.bool import FALSE, TRUE
    from starkware.cairo.common.cairo_builtins import EcOpBuiltin, SignatureBuiltin
    from starkware.cairo.common.ec import StarkCurve, ec_add, ec_mul, ec_sub, is_x_on_curve, recover_y
    from starkware.cairo.common.ec_point import EcPoint
    
    // Verifies that the prover knows a signature of the given public_key on the given message.
    // Prover assumption: (signature_r, signature_s) is a valid signature for the given public_key
    // on the given message.
    func verify_ecdsa_signature{ecdsa_ptr: SignatureBuiltin*}(
        message, public_key, signature_r, signature_s
    ) {
        %{ ecdsa_builtin.add_signature(ids.ecdsa_ptr.address_, (ids.signature_r, ids.signature_s)) %}
        assert ecdsa_ptr.message = message;
        assert ecdsa_ptr.pub_key = public_key;
    
        let ecdsa_ptr = ecdsa_ptr + SignatureBuiltin.SIZE;
        return ();
    }
    
    // Checks if (signature_r, signature_s) is a valid signature for the given public_key
    // on the given message.
    // Arguments:
    //   message - the signed message.
    //   public_key - the public key corresponding to the key with which the message was signed.
    //   signature_r - the r component of the ECDSA signature.
    //   signature_s - the s component of the ECDSA signature.
    // Returns:
    //   res - TRUE if the signature is valid, FALSE otherwise.
    func check_ecdsa_signature{ec_op_ptr: EcOpBuiltin*}(
        message, public_key, signature_r, signature_s
    ) -> (res: felt) {
        alloc_locals;
        // Check that s != 0 (mod StarkCurve.ORDER).
        if (signature_s == 0) {
            return (res=FALSE);
        }
        if (signature_s == StarkCurve.ORDER) {
            return (res=FALSE);
        }
        if (signature_r == StarkCurve.ORDER) {
            return (res=FALSE);
        }
    
        // Check that the public key is the x coordinate of a point on the curve.
        let on_curve: felt = is_x_on_curve(public_key);
        if (on_curve == FALSE) {
            return (res=FALSE);
        }
        // Check that r is the x coordinate of a point on the curve.
        // Note that this ensures that r != 0.
        let on_curve: felt = is_x_on_curve(signature_r);
        if (on_curve == FALSE) {
            return (res=FALSE);
        }
    
        // To verify ECDSA, obtain:
        //   zG = z * G, where z is the message and G is a generator of the EC.
        //   rQ = r * Q, where Q.x = public_key.
        //   sR = s * R, where R.x = r.
        // and check that:
        //   zG +/- rQ = +/- sR, or more efficiently that:
        //   (zG +/- rQ).x = sR.x.
        let (zG: EcPoint) = ec_mul(m=message, p=EcPoint(x=StarkCurve.GEN_X, y=StarkCurve.GEN_Y));
        let (public_key_point: EcPoint) = recover_y(public_key);
        let (rQ: EcPoint) = ec_mul(signature_r, public_key_point);
        let (signature_r_point: EcPoint) = recover_y(signature_r);
        let (sR: EcPoint) = ec_mul(signature_s, signature_r_point);
    
        let (candidate: EcPoint) = ec_add(zG, rQ);
        if (candidate.x == sR.x) {
            return (res=TRUE);
        }
    
        let (candidate: EcPoint) = ec_sub(zG, rQ);
        if (candidate.x == sR.x) {
            return (res=TRUE);
        }
    
        return (res=FALSE);
    }
    ```
    
- **bitwise**
    
    This structure defines the memory organization for the bitwise built-in. It's utilized by functions in the common library that leverage the bitwise built-in functionality. For example, in the bitwise_xor() function, there's an implicit argument of type BitwiseBuiltin*, which internally manages the next available built-in instance.
    
    The struct consists of members of type felt:
    
    - **x**: Represents the first operand.
    - **y**: Represents the second operand.
    - **x_and_y**: Holds the result of the bitwise AND operation between x and y.
    - **x_xor_y**: Stores the result of the bitwise XOR operation between x and y.
    - **x_or_y**: Holds the result of the bitwise OR operation between x and y.
    
    Additionally, there's a pointer named bitwise_ptr, which is of type BitwiseBuiltin*. This pointer facilitates internal management of the bitwise built-in instances.
    
    ```rust
    from starkware.cairo.common.cairo_builtins import BitwiseBuiltin
    
    const ALL_ONES = -1;
    
    // Computes the bitwise operations and, xor and or.
    // Arguments:
    //   bitwise_ptr - the bitwise builtin pointer.
    //   x, y - the two field elements to operate on, in this order. Both inputs should be 251-bit
    //     integers, and are taken as unsigned ints.
    // Returns:
    //   x_and_y = x & y (bitwise and).
    //   x_xor_y = x ^ y (bitwise xor).
    //   x_or_y = x | y (bitwise or).
    func bitwise_operations{bitwise_ptr: BitwiseBuiltin*}(x: felt, y: felt) -> (
        x_and_y: felt, x_xor_y: felt, x_or_y: felt
    ) {
        bitwise_ptr.x = x;
        bitwise_ptr.y = y;
        let x_and_y = bitwise_ptr.x_and_y;
        let x_xor_y = bitwise_ptr.x_xor_y;
        let x_or_y = bitwise_ptr.x_or_y;
        let bitwise_ptr = bitwise_ptr + BitwiseBuiltin.SIZE;
        return (x_and_y=x_and_y, x_xor_y=x_xor_y, x_or_y=x_or_y);
    }
    
    // Computes the bitwise and of two inputs.
    // Arguments:
    //   bitwise_ptr - the bitwise builtin pointer.
    //   x, y - the two field elements to operate on, in this order. Both inputs should be 251-bit
    //     integers, and are taken as unsigned ints.
    // Returns:
    //   x_and_y = x & y (bitwise and).
    func bitwise_and{bitwise_ptr: BitwiseBuiltin*}(x: felt, y: felt) -> (x_and_y: felt) {
        bitwise_ptr.x = x;
        bitwise_ptr.y = y;
        let x_and_y = bitwise_ptr.x_and_y;
        let x_xor_y = bitwise_ptr.x_xor_y;
        let x_or_y = bitwise_ptr.x_or_y;
        let bitwise_ptr = bitwise_ptr + BitwiseBuiltin.SIZE;
        return (x_and_y=x_and_y);
    }
    
    // Computes the bitwise xor of two inputs.
    // Arguments:
    //   bitwise_ptr - the bitwise builtin pointer.
    //   x, y - the two field elements to operate on, in this order. Both inputs should be 251-bit
    //     integers, and are taken as unsigned ints.
    // Returns:
    //   x_xor_y = x ^ y (bitwise xor).
    func bitwise_xor{bitwise_ptr: BitwiseBuiltin*}(x: felt, y: felt) -> (x_xor_y: felt) {
        bitwise_ptr.x = x;
        bitwise_ptr.y = y;
        let x_and_y = bitwise_ptr.x_and_y;
        let x_xor_y = bitwise_ptr.x_xor_y;
        let x_or_y = bitwise_ptr.x_or_y;
        let bitwise_ptr = bitwise_ptr + BitwiseBuiltin.SIZE;
        return (x_xor_y=x_xor_y);
    }
    
    // Computes the bitwise or of two inputs.
    // Arguments:
    //   bitwise_ptr - the bitwise builtin pointer.
    //   x, y - the two field elements to operate on, in this order. Both inputs should be 251-bit
    //     integers, and are taken as unsigned ints.
    // Returns:
    //   x_or_y = x | y (bitwise or).
    func bitwise_or{bitwise_ptr: BitwiseBuiltin*}(x: felt, y: felt) -> (x_or_y: felt) {
        bitwise_ptr.x = x;
        bitwise_ptr.y = y;
        let x_and_y = bitwise_ptr.x_and_y;
        let x_xor_y = bitwise_ptr.x_xor_y;
        let x_or_y = bitwise_ptr.x_or_y;
        let bitwise_ptr = bitwise_ptr + BitwiseBuiltin.SIZE;
        return (x_or_y=x_or_y);
    }
    
    // Computes the bitwise not of a single 251-bit integer.
    // Argument:
    //   x - the field element to operate on. The input should be a 251-bit
    //     integer, and is taken as unsigned int.
    // Returns:
    //   not_x = ~x (bitwise not).
    func bitwise_not(x: felt) -> (not_x: felt) {
        return (not_x=ALL_ONES - x);
    }
    ```