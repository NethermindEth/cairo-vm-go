## Implementing new Cairo 0 hints
If you want to implement a new Cairo zero hint you'll have to follow the next steps. Let's use "Assert250bits" hint as an example.

1- As we are only allowing whitelisted hints, a hint can be recognized by its code. So the first step would be to add the code string in the [hintcode.go](hintcode.go) file. Ensure that the variable naming convention follows the established pattern in the file, using `<hintName>Code` as the format.
```
assert250bits string = "from starkware.cairo.common.math_utils import as_int\n\n# Correctness check.\nvalue = as_int(ids.value, PRIME) % PRIME\nassert value < ids.UPPER_BOUND, f'{value} is outside of the range [0, 2**250).'\n\n# Calculation for the assertion.\nids.high, ids.low = divmod(ids.value, ids.SHIFT)"
```
Also notice that the hints are grouped together by functionality. The code of each hint can be found in the [cairo-lang library](https://github.com/starkware-libs/cairo-lang/tree/master/src/starkware/cairo/common) or directly in the VM in Go by LambdaClass where they [gathered all hints](https://github.com/lambdaclass/cairo-vm_in_go/tree/main/pkg/hints/hint_codes)

2- Update the `GetHintFromCode` method within the [zerohint.go](zerohint.go) file by adding the new hint to the switch-case structure.
```
    switch rawHint.Code {
    // ...
    case assert250bits:
        return createAssert250bitsHinter(resolver)
    // ...

```
The method to handle your specific hint should be named `create<HintName>Hinter` as in the example above.

3- This method takes a `hintReferenceResolver` as a parameter. This structure contains a map of `(string, Reference)` pairs which stores the name of the variable involved in the hint with the corresponding operand. The idea is to get all operanders involved in the hint and call the constructor of the specific hint.
<!-- TODO: Add some documentation about operands -->

We also group the implementation of the hints in different files according to their functionality, so the definition of this method should go in the corresponding file. In the case of `Assert250bits` it goes in [zerohint_math.go](zerohint_math.go).

4- Now the structure we should return has to implement the interface:

```
type Hinter interface {
	fmt.Stringer

	Execute(vm *VM.VirtualMachine, ctx *HintRunnerContext) error
}
```

so to make this process easier we've created a generic structure `GenericZeroHinter` that will allow you to pass on all the information needed. It will look something like this:

```
func newAssert250bitsHint(low, high, value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Assert250bits",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			// Implementation of the hint goes here
		},
	}
}
```

5- This is the part that gets insteresting, because implementing the hint means you have to write a code in go that has the same behavior as the code of the hint in python. The difficulty of course depends on the hint. Some things that are good to know here:

- It might be the case that the hint is already implemented as part of the **core** pkg (Cairo 1 hints), so you might want to check first if it is. Such is the case for `AssertLeFelt` hint. In those cases you would only need to return the corresponding hint structure from the **core** pkg.
- The operanders allow you to get the address in memory of the variable you need to use. When this variable is a `felt`, simply reading the value from that address is sufficient. However, if the variable is a `struct`, you need to retrieve several consecutive values from memory. For this, you can use the `GetConsecutiveMemoryValues(addr MemoryAddress, size int16)` method. 
For example, in the `uint256add` hint, you need to add `a` and `b`, which are both `Uint256` structs. To achieve this, we use the `GetUint256AsFelts` utility, which calls `ReadFromAddress` twice to read the low and high parts of the `uint256` in memory. It's important to note that `GetUint256AsFelts` does not use `GetConsecutiveMemoryValues`; instead, it directly uses `ReadFromAddress` two times. Both `GetUint256AsFelts` and `GetConsecutiveMemoryValues` utilize `ReadFromAddress`, but in different ways: the former calls it twice directly, while the latter uses it inside a loop to read multiple consecutive values.
- When there is an assignment to a variable that has been defined in the cairo code not inside the hint code (you can recognize it by the name being used `ids.<variable_name>`), then the assignment should be translated into a write to memory to the corresponding address of the variable using `func (memory *Memory) WriteToAddress(address *MemoryAddress, value *MemoryValue) error` from the `vm/memory` package.
- It might happen that a variable used in the hint code wasn't defined there nor in the cairo code (as in the previous point). This means the variable was defined in a previous hint in the same execution scope. You might want to check [Execution Scopes](#execution-scopes) section for better understanding.
 
6- After the implementation is completed, you'll need to add unit tests to check the behavior of your code is correct. Make sure to add your tests in the corresponding file.

<!-- TODO: Add some documentation on Unit Testing for Cairo 0 hints -->

## Execution Scopes for hints

Let's use as an example [assert_le_felt](https://github.com/starkware-libs/cairo-lang/tree/master/src/starkware/cairo/common) function:
```
@known_ap_change
func assert_le_felt{range_check_ptr}(a, b) {
    // ceil(PRIME / 3 / 2 ** 128).
    const PRIME_OVER_3_HIGH = 0x2aaaaaaaaaaaab05555555555555556;
    // ceil(PRIME / 2 / 2 ** 128).
    const PRIME_OVER_2_HIGH = 0x4000000000000088000000000000001;
    // The numbers [0, a, b, PRIME - 1] should be ordered. To prove that, we show that two of the
    // 3 arcs {0 -> a, a -> b, b -> PRIME - 1} are small:
    //   One is less than PRIME / 3 + 2 ** 129.
    //   Another is less than PRIME / 2 + 2 ** 129.
    // Since the sum of the lengths of these two arcs is less than PRIME, there is no wrap-around.
    %{
        import itertools

        from starkware.cairo.common.math_utils import assert_integer
        assert_integer(ids.a)
        assert_integer(ids.b)
        a = ids.a % PRIME
        b = ids.b % PRIME
        assert a <= b, f'a = {a} is not less than or equal to b = {b}.'

        # Find an arc less than PRIME / 3, and another less than PRIME / 2.
        lengths_and_indices = [(a, 0), (b - a, 1), (PRIME - 1 - b, 2)]
        lengths_and_indices.sort()
        assert lengths_and_indices[0][0] <= PRIME // 3 and lengths_and_indices[1][0] <= PRIME // 2
        excluded = lengths_and_indices[2][1]

        memory[ids.range_check_ptr + 1], memory[ids.range_check_ptr + 0] = (
            divmod(lengths_and_indices[0][0], ids.PRIME_OVER_3_HIGH))
        memory[ids.range_check_ptr + 3], memory[ids.range_check_ptr + 2] = (
            divmod(lengths_and_indices[1][0], ids.PRIME_OVER_2_HIGH))
    %}
    // Guess two arc lengths.
    tempvar arc_short = [range_check_ptr] + [range_check_ptr + 1] * PRIME_OVER_3_HIGH;
    tempvar arc_long = [range_check_ptr + 2] + [range_check_ptr + 3] * PRIME_OVER_2_HIGH;
    let range_check_ptr = range_check_ptr + 4;

    // First, choose which arc to exclude from {0 -> a, a -> b, b -> PRIME - 1}.
    // Then, to compare the set of two arc lengths, compare their sum and product.
    let arc_sum = arc_short + arc_long;
    let arc_prod = arc_short * arc_long;

    // Exclude "0 -> a".
    %{ memory[ap] = 1 if excluded != 0 else 0 %}
    jmp skip_exclude_a if [ap] != 0, ap++;
    assert arc_sum = (-1) - a;
    assert arc_prod = (a - b) * (1 + b);
    return ();

    // Exclude "a -> b".
    skip_exclude_a:
    %{ memory[ap] = 1 if excluded != 1 else 0 %}
    jmp skip_exclude_b_minus_a if [ap] != 0, ap++;
    tempvar m1mb = (-1) - b;
    assert arc_sum = a + m1mb;
    assert arc_prod = a * m1mb;
    return ();

    // Exclude "b -> PRIME - 1".
    skip_exclude_b_minus_a:
    %{ assert excluded == 2 %}
    assert arc_sum = b;
    assert arc_prod = a * (b - a);
    ap += 2;
    return ();
}
```

There are four hint blocks in this code. Notice the `excluded` variable is defined in the first one, but it's actually being used in the following three. So just by having the code of one hint isn't enough for executing it, even though all of them belong to the same scope.

The current solution uses the `HintRunnerContext` structure, which is passed down to the implementation of each hint. This structure contains a `ScopeManager` that will handle all the operations related to scope, such as:
- Creating a new scope: Use method `EnterScope()` when the hint code uses `vm_enter_scope()` method.
- Exiting current scope: Use method `ExitScope()` when the hint uses `vm_exit_scope()` method.
- Variable declaration and assingment: Just like the `excluded` variable in the first hint block of the previous example. Use `AssignVariable()` method.
- Accessing variable values: Just like the `excluded` variable in the last three hint blocks of the previous example. Use `GetVariableValue()` method.

Check [scope.go](../hinter/scope.go) file for details in the implementation.