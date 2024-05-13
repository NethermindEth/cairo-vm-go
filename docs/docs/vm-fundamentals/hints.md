---
sidebar_position: 3
---


# Hint
  ## HINTS  IN CAIRO VIRTUAL MACHINE
Hints in Cairo are pieces of Python code only executed by the sequencer.

Hints are used to help the Virtual Machine determine the next path of execution it should take.

Below are some Cairo code with hints:

```cairo
      [ap] = 25, ap++;
    %{
        import math
         memory[ap] = int(math.sqrt(memory[ap - 1]))
     %}
      [ap - 1] = [ap] * [ap], ap++;
```


```cairo
      func main() {
          %{ memory[ap] = program_input['secret'] %}
         [ap] = [ap], ap++;
         ret;
      }
``` 



### WHY HINTS RUN ON SEQUENCERS/ PROVERS AND NOT VERIFIERS

Sequencers are the leads that produces blocks by executing transaction and updating the blockchain state.
During the proof generation process, the prover uses hints to generate additional constraints or assumptions about the program's behavior, which makes part of the proof sent to the verifier for verification. In the Cairo Virtual Machine, hints are sorted out during different stages of the compilation and execution process by the sequencer. However, they are not directly involved in the verification stage. Hints are primarily for performance optimization.
    
The key is the implementation of relevant and accurate hints that assist the prover or sequencers in their respective tasks.



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

## IMPLEMENTING HINTS IN CAIRO VIRTUAL MACHINE
     
 In Cairo Virtual Machine, hints are accepted and recognized by its code. 
Hints are implemented using the *%{...%}* syntax, which allows you to insert Python code that is only executed by the prover. When implementing hints, the *'variable naming convention'* should be considered. The method to handle specific hints should be outlined like this; *'create hintName Hinter'*. Hints should be grouped by functionality in Cairo programs. The structure returned should implement the interface. Implementing the hint means you have to write code in Go that has the same behavior as the code of the hint in Python. Unit tests are added to check if the behavior of your code is correct.

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



