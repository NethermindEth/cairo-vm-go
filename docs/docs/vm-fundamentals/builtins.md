---
sidebar_position: 2
---

# Builtins

Regular Cairo instructions are like regular Legos, they work, but for complex things they can be slow. Builtins are special Lego pieces that snap together easily, like pre-built wheels or doors. This makes building things much faster. In a Cairo context, builtins help write clean and efficient Cairo programs, just like these special Lego pieces help build cool things faster.

This section provides an overview of the builtins functions available in the Cairo Virtual Machine (Cairo VM).

### What is builtins in cairo?

Builtins are predefined optimized low-level execution units which are added to the Cairo CPU board to perform predefined computations which are expensive to perform in Cairo.

Imagine the Cairo VM as a powerful computer. Regular Cairo instructions are like the basic building blocks the computer can understand to perform calculations, but for some complex tasks, these building blocks can be slow to put together. Builtins are like special hardware additions to the CPU that allow the computer to perform specific tasks much faster. Just like a computer with a powerful graphics card runs games smoother, builtins help write efficient Cairo programs.

### Builtins and Cairo memory

**Builtins act as special extensions to the Cairo VM's memory, enforcing specific rules on designated areas.** 

Imagine the VM memory as a filing cabinet, builtins are like add-on modules that create special folders within the cabinet and each builtin gets its own folder and dictates the rules for what can be stored there. For instance, a "range-check" builtin might create a folder where all values must be between 0 and a very large number.

**Communication between the CPU and builtins happens entirely through this memory.** 

The CPU doesn't directly call builtins, instead, it treats the builtins memory folder like a special device to use a builtin. A function gets a pointer to the builtins folder as an argument, the function then reads or writes data to specific cells within the folder, following the rules set by the builtin. This method is similar to how external devices interact with a computer, they don't talk directly to the CPU but use the memory space for communication.

**Adding builtins doesn't require modifying the CPU itself.** 

Both the CPU and builtins share the same memory space, making them efficient tools. Additionally, each function using a builtin receives a pointer and is responsible for returning a pointer to the next unused cell within the builtin's memory area. This ensures proper management of space allocated for each builtin instance.

### General types of builtins in Cairo VM

- **Mathematical builtins.**
    
    These perform various mathematical operations, potentially faster than using regular Cairo instructions. Examples include addition, subtraction, multiplication, bitwise operations (AND, OR, XOR, etc.), and some even handle more complex operations like modular exponentiation.
    
- **Cryptographic builtins.**
    
    These handle cryptographic functions used in zk-STARK proofs. Examples include functions for hashing (like Keccak) and various elliptic curve cryptography operations. Some examples are:
    
    **→ Keccak hashing:**
    
    This builtin performs the Keccak hash function, a one-way function that creates a unique and fixed-size output from any input data. It's used to compress data and ensure its integrity in zk-STARK proofs.
    
    **→ Scalar multiplication:**
    
    This builtin performs point multiplication on elliptic curves, a fundamental operation in elliptic curve cryptography (ECC). ECC is used for secure communication and signature verification in zk-STARKs.
    
- **Memory management builtins.**
    
    These help manage memory allocation and deallocation within a Cairo program. This can be crucial for optimizing memory usage. Some examples are:
    
    **→ Memory allocation:**
    
    This builtin allocates a specific amount of memory within the Cairo VM for a program to use. This ensures the program doesn't try to access non-existent memory locations.
    
    **→ Memory deallocation:**
    
    This builtin frees up memory that is no longer needed by a program. This prevents memory leaks and optimizes memory usage for subsequent computations.
    
- **Comparison builtins.**
    
    These allow for comparisons between values, like checking if one value is greater than another. Some examples are:
    
    **→ Equal to:**
    
    This builtin checks if two felt values are identical. This is useful for conditional statements within a program.
    
    **→ Greater than:**
    
    This builtin checks if one felt value is greater than another. This allows for branching logic based on the relative size of values.
    

### Specific builtins in Cairo VM

- **output**
    
    The output builtin, accessed with a pointer to type felt, is used for writing program outputs.
    
    ```rust
    // Function to add two numbers and write the sum to the output
    func add(a felt, b felt) -> (felt) {
      let sum = a + b
      // Use the output builtin to write the sum to the program's output
      output(sum)
      return sum
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
    

### How bulitins perform complex operations at the cost of extra constraints when proving

While built-in functions offer convenience and efficiency, they may introduce constraints or limitations during testing and verification processes that are primarily related to security, correctness, and resource management. Since built-in functions often handle critical operations such as cryptographic hashing or signature verification, it's essential to ensure their correctness and robustness under various scenarios.

Let's delve into how built-in functions achieve this and the constraints they impose.

- **Constraints on Resource Usage.**
    
    Built-in functions often consume computational resources, such as CPU cycles and memory, to perform computations or execute algorithms. Efficient utilization of computational resources is essential for smart contracts to ensure timely execution and minimize transaction costs. However, excessive computational resource consumption can lead to performance degradation or denial-of-service vulnerabilities, where malicious actors exploit resource-intensive operations to disrupt contract execution.
    
    - **Testing Strategies for Resource Management.**
        
        To address resource usage constraints effectively, developers can employ various testing strategies focused on resource management:
        
        - **Stress Testing.**
            
            Stress testing involves subjecting the system to extreme workload conditions to evaluate its performance and resilience under high resource utilization. By pushing the system to its limits, stress tests help identify potential failure points or scalability limitations and enable developers to optimize resource management strategies accordingly.
            
        - **Resource Profiling.**
            
            Resource profiling involves analyzing the resource consumption patterns of built-in functions to identify areas for optimization. By profiling memory usage, CPU utilization, and other resource metrics, developers can pinpoint inefficiencies in resource management and prioritize optimization efforts to improve overall system performance.
            
        - **Benchmarking.**
            
            Benchmarking compares the performance of built-in functions against established benchmarks or performance targets to gauge their efficiency and effectiveness. By benchmarking resource usage metrics such as memory footprint and execution time, developers can assess the relative performance of different implementations and identify opportunities for optimization.
            
- **Constraints on Security.**
    
    Security is paramount in smart contract development, particularly in blockchain environments where transactions are irreversible and funds are at stake. Built-in functions must be thoroughly tested to identify and mitigate security vulnerabilities, such as integer overflow/underflow, reentrancy attacks, or unintended access control issues. Testing for security vulnerabilities often involves techniques like fuzz testing, static analysis, and formal verification to ensure robustness against potential attacks.
    
    - **Types of Security Constraints.**
        - **Integer Overflow/Underflow.**
            
            Built-in functions that involve arithmetic operations must guard against integer overflow and underflow vulnerabilities. These vulnerabilities occur when the result of an arithmetic operation exceeds the maximum or minimum representable value for the data type, potentially leading to unintended behavior or manipulation of contract state. 
            
        - **Reentrancy Attacks.**
            
            Reentrancy attacks exploit the asynchronous nature of smart contract execution to manipulate contract state in unintended ways. This vulnerability arises when a contract interacts with external contracts or sends funds before completing its internal state changes, allowing an attacker to re-enter the contract and perform additional operations before the previous ones are finalized. 
            
        - **Access Control Issues**: Built-in functions that govern access control within smart contracts must enforce permissions rigorously to prevent unauthorized actions or privilege escalation. Security vulnerabilities may arise if access control mechanisms are improperly implemented or bypassed, allowing unauthorized users to modify contract state or execute privileged operations.
    - **Testing Strategies for Security**
        - **Fuzz Testing.**
            
            Fuzz testing involves generating a large volume of random or invalid inputs to uncover security vulnerabilities, including those related to built-in functions. By subjecting built-in functions to diverse input conditions, fuzz testing helps identify potential edge cases and corner cases that may trigger security vulnerabilities, such as buffer overflows or unexpected behavior.
            
        - **Static Analysis.**
            
            Static analysis tools analyze the source code of smart contracts to identify potential security vulnerabilities, including those related to built-in functions. These tools can detect common coding patterns associated with security vulnerabilities, such as unchecked external calls or improper input validation, helping developers identify and mitigate risks before deployment.
            
        - **Formal Verification.**
            
            Formal verification techniques provide mathematical assurance of the correctness and security properties of built-in functions by rigorously analyzing their behavior against formal specifications or security properties. By employing techniques such as theorem proving or model checking, formal verification can help ensure that built-in functions adhere to security best practices and mitigate potential vulnerabilities.
            
- **Constraints on Compatibility**
    
     Smart contracts often interact with other contracts or external systems, requiring compatibility with specific interfaces or protocols. Built-in functions must adhere to these compatibility constraints to ensure seamless integration with other components of the blockchain ecosystem. Testing for compatibility involves verifying that built-in functions comply with relevant standards and specifications, such as ERC (Ethereum Request for Comment) proposals for Ethereum-based contracts.
    
    - **Interoperability Requirements.**
        
        Smart contracts frequently interact with other contracts, decentralized applications (dApps), or external systems, necessitating compatibility with various interfaces, protocols, and standards. Built-in functions must adhere to these interoperability requirements to ensure smooth communication and collaboration within the blockchain ecosystem. Incompatibilities can lead to functionality errors, data inconsistencies, or even contract failures.
        
    - **Compatibility Testing Strategies.**
        - **Conformance Testing.**
            
            Conformance testing verifies that built-in functions adhere to relevant standards and specifications, such as ERC proposals for Ethereum-based contracts. Test cases are designed to validate compliance with specific interface definitions, parameter formats, and behavior expectations outlined in the standards. By ensuring conformance to established standards, compatibility with existing infrastructure and ecosystem components is maintained.
            
        - **Integration Testing.**
            
            Integration testing evaluates the interaction between built-in functions and other components of the blockchain ecosystem, including external contracts, dApps, or oracle services. Test scenarios simulate real-world usage scenarios to assess compatibility, data exchange, and interoperability. By validating seamless integration with diverse ecosystem components, integration testing ensures that built-in functions perform reliably across different environments.
            
        - **Protocol Compatibility Testing.**
            
            Protocol compatibility testing assesses the compatibility of built-in functions with underlying blockchain protocols and network upgrades. Test cases evaluate functionality, performance, and data integrity under various protocol versions or network configurations. By ensuring compatibility with evolving blockchain protocols, built-in functions remain compatible with future platform upgrades and enhancements.
            

### Conclusion

Builtins in Cairo are predefined optimized low-level execution units that perform specific computations more efficiently. They act as special extensions to the Cairo VM's memory and do not require modifications to the CPU. Builtins include mathematical, cryptographic, memory management, and comparison types. They can perform complex operations but may introduce constraints during testing and verification processes related to resource usage, security, and compatibility. Developers can employ various testing strategies to manage these constraints effectively, ensuring the robustness and efficiency of Cairo programs.