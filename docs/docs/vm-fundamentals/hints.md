---
sidebar_position: 3
---

# Hint
  ## HINTS  IN CAIRO VIRTUAL MACHINE
Hints in Cairo are pieces of Python code only seen and executed by the prover and are not included in the bytecode. They instruct the prover on how to handle nondeterministic instructions. These nondeterministic instructions are programs that have different outcomes at execution.
    
Hints are guides that developers can provide to the Cairo virtual machine to optimize execution or provide additional information about the program. These hints help improve performance or help the VM make better decisions during the execution of programs. 

## TYPES OF HINTS IN CARIO VIRTUAL MACHINE
The hints in Cairo Virtual Machine that influence the behavior of programs include:
### EXECUTION HINTS
 they guide the execution of cairo programs. They include guides on loop unrolling, parallel execution, or optimizations to enhance. It influences how the VM schedules and executes instructions, thereby improving performance.
In Cairo, execution hints use the *hint* keyword, followed by a *hint name* and *a value*.

 @example:

 ```cairo  
      func main() -> felt*:
      // some code here
      hint memory_behavior = MemoryBehavior.Sequential;
      // more code here
  ```
  The memory_behavior hint informs the Cairo compiler that the code following the hint   
 sequentially accesses memory.

 @example:

 ``` cairo
     func main() {
         // example of loop unrolling hint
         @loop_unroll(4)
         for i = 0 to 10 {
         // Loop body
        // example of parallel execution hint
         }
         @parallel_execution
         {
         // Code block for execution in parallel
         }
     }
 ```

### MEMORY HINTS
 These hints are memory aids that provide additional information to the prover on computing values in the Cairo program. They are functional when provers can't handle certain expressions implicitly.

@example:

``` cairo
       [ap] = 25, ap++;
     %{
        import math
        memory[ap] = int(math.sqrt(memory[ap - 1]))
     %}
        [ap - 1] = [ap] * [ap], ap++;
```

In this example, the hint *%{ import math memory[ap] = int(math.sqrt(memory[ap - 1])) %}* is added before the instruction *[ap - 1] = [ap] \* [ap], ap++;*. The hint calculates the square root of the value stored in *memory[ap - 1]* and stores the result in *memory[ap]*.

### RESOURCE MANAGEMENT HINTS
These hints provide additional information on managing resources like memory and stack space. They help the virtual Machine optimize resource usage and prevent resource leaks. Resource management hints need to be applied judiciously and only when necessary.

@example:

 ``` cairo
    %{__resource_management_hint__ memory_pages_allocated = 10 %}
        [ap] = 0, ap++;
     for i in range(10):
        [ap] = i, ap++;
   %{__resource_management_hint__ memory_pages_allocated =  %}
     for i in range(5):
      [ap] = i * i, ap++;
 ```

The first hint will ensure the program has *`10`* memory pages to execute the first loop that stores *`10`* integers on the stack.
The second hint will ensure the program does not use more memory than necessary.

### DEBUGGING HINTS
They aid in troubleshooting and debugging Cairo programs. They identify and diagnose issues in the program. It debugs errors related to memory access and manipulation. Debugging hints are deleted from the program once the debugging is complete.

 @example:

```cairo
    %{
       import os
       print("Debugging hint: Memory at address {}".format(memory[ap - 1]))
    %}
    [ap - 1] = 42, ap++;
```

The debugging hint prints the value stored in memory and the address stored in the register *ap - 1*.

### SECURITY HINTS
Security hints in Cairo VM provide security features and protections for programs through directions for input validation, access control, or encryption to safeguard against potential vulnerabilities and threats.

@example:

```cairo
     %{ 
       import cairo 
        cairo.set_max_stack_size(1024) 
     %} 
       [ap] = 42, ap++;
 ```
The security hint sets the maximum stack size for the program to *1024*.

## HOW HINTS WORK IN CAIRO VIRTUAL MACHINE

1. **Execution Trace**: When executing a smart contract( a digital agreement signed and stored on a blockchain network, which executes automatically when the contract's terms and conditions (T&C) are satisfied) written in Cairo, the VM generates an execution trace. This trace contains information about the sequence of instructions executed, the values of variables at different points, and the instructions or function calls of an important program that are executed or evaluated.

2. **Hinting Mechanism**: Cairo VM gives a procedure for developers to provide hints to the verification tool. These hints can come in various forms, such as assertions, assumptions, or annotations within the Cairo code.

3. **Integration with Verification Tools**: Cairo VM integrates with external verification tools that analyze the execution traces and verify the features of the smart contracts. These tools can include assumptions provers, model checkers, or symbolic execution engines.

4. **Guiding Verification with Hints**: During the verification process, the hints provided by developers serve as guides for analysis performed by the verification tools.

For example:
 - **Assertion Checking**: Developers or testers use specific statements, known as assertions, to validate the expected   behavior of a program or system. Verification tools check whether these assertions fail during the analysis.
    
 - **Invariant Detection**: Hints may guide the tool to identify the uniformity of loop iteration or other program properties responsible for the accuracy of the contract. The tool then attempts to prove these properties using the execution trace.

 - **Path Exploration**: Hints can help the tools explore specific paths or branches of the program that are considered fit for verification. It helps in focusing the analysis on relevant parts of the code.
   

5. **Feedback Loop**: As the verification tool analyzes the execution trace and attempts to prove properties guided by the hints, it may encounter challenges or limitations. In such cases, feedback mechanisms allow developers to refine their hints or modify the code to improve the chances of successful verification.

### WHY HINTS RUN ON SEQUENCERS/ PROVERS AND NOT VERIFIERS

The verifier in Cairo is responsible for ensuring that a program attaches safety and accuracy features. It checks that the program satisfies limitations such as memory safety, type safety, and functional correctness.

Sequencers in Cairo are responsible for organizing the execution of transactions, managing instruction ordering and dependencies, exploiting multiple processors, and ensuring the integrity of the program's state throughout the execution process.   
     
Provers attempt to mathematically prove the correctness of a program concerning a given specification. The prover uses these hints to access the program's behavior and verify that it satisfies specified properties and constraints. In this context, hints could guide the prover towards particular properties or parts of the codebase that are relevant to the proof.

During the proof generation process, the prover uses hints to generate additional constraints or assumptions about the program's behavior, which makes part of the proof sent to the verifier for verification. In the Cairo VM, hints are sorted out during different stages of the compilation and execution process by the sequencer and the prover. However, they are not directly involved in the verification stage. Hints are primarily for performance optimization.
    
The key is the implementation of relevant and accurate hints that assist the prover or sequencers in their respective tasks.

Example:
The 'random' instruction generates a random number, which can be different each time of instruction execution. The verifier cannot implement these instructions, as it would require access to nondeterministic information, which would beat the purpose of the zero-knowledge proof.
To solve this problem, Cairo uses hints to instruct the prover on how to handle nondeterministic instructions. The prover will execute the nondeterministic instructions and provide the verifier with enough information to check whether or not the program is accurate without revealing any sensitive information.
Example:
Let's consider a Cairo function that calculates the factorial of a given input *`n`*:

```Cairo
    @public
   func factorial(n : felt) -> (result : felt):
    if n <= 1:
        result := 1
    else:
     result := n * factorial(n - 1)
```

In this code, the hint *@public* indicates that the *`n`* factorial should be publicly available in the Cairo program.
The provers/sequencers will utilize the hint to generate a proof or sequence of instructions for the function to ensure maximum optimization. The verifier will check that the generated proof or sequence of instructions has correctly calculated the factorial of the input *`n`* or any other specification is satisfied.

## IMPLEMENTATING HINTS IN CAIRO VIRTUAL MACHINE

IMPLEMENTING HINTS IN CAIRO VIRTUAL MACHINE
      
 In Cairo Virtual Machine, hints are accepted and recognized by its code. 
Hints are implemented using the *`%{....%}`* syntax, which allows you to insert Python code that is only executed by the prover. When implementing hints, the *variable naming convention* should be considered. The method to handle specific hints should be outlined like this *create<hintName>Hinter*. Hints should be grouped by functionality in Cairo programs. The structure returned should implement the interface. Implementing the hint means you have to write code in Go that has the same behavior as the code of the hint in Python. Unit tests are added to check if the behavior of your code is correct.

Examples of how to implement a hint in cairo:

``` cairo
      from starkware.cairo.common.hints import Hint

        class FactorialHint(Hint):
        def __init__(self, n):
            self.n = n

        def process(self, cairo_ctx):
           result = 1
        for i in range(2, self.n + 1):
            result *= i
        return [result]

    //Usage   
      [ap] = 5, ap++;
    %{
       hint_processor = FactorialHint(memory[ap - 1])
        memory[ap] = cairo_runner.run_from_entrypoint(
           entrypoint,
           [2, (2,0)],
           False,
           vm,
           hint_processor
        )[0];
    %}
      [ap - 1] = [ap] * [ap], ap++;

```

```cairo
   from starkware.cairo.common.math import assert_nn_le

  struct KeyValue {
      key: felt,
      value: felt,
    }

       struct HintResult {
       idx: felt,
    }

    func get_value_by_key{range_check_ptr}(
        list: KeyValue*, size, key
     ) -> (value: felt) {
       alloc_locals;
       local idx;
       local hint_result: HintResult;

    %{
        # Define the hint function
        def hint_get_index_by_key(list, size, key):
            for i in range(size):
                if list[i].key == key:
                    return HintResult(i)
            raise ValueError("Key not found")

        # Call the hint function and store the result in memory
        hint_result = hint_get_index_by_key(list, size, key)
        memory[ap] = hint_result.idx
        ap += 1
    %}

       # Load the result of the hint function from memory
       hint_result = memory[ap - 1];
         ap -= 1;

       assert_nn_le(hint_result.idx, size);
       return list[hint_result.idx].value;
    }

```



