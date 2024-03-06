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
- The operanders allow you to get the address in memory of the variable you need to use. When this variable is a felt just reading the value from that address is enough, but sometimes the variable is actually a struct so you need to get several consecutive values in memory, for that you can use `GetConsecutiveValues(vm *VM.VirtualMachine, ref ResOperander, size int16)` method. For example in `uint256add` hint you need to add `a` and `b` which are both `Uint256` structs, so we used `GetUint256AsFelts` utility which basically calls that function with size 2.
- When there is an assignment in the hint code to a variable that has been defined outside of its scope, then the assignment should be translated into a write to memory to the corresponding address of the variable using `func (memory *Memory) WriteToAddress(address *MemoryAddress, value *MemoryValue) error` from the `vm/memory` package.
  
6- After the implementation is completed, then you'll need to add unit tests to check the behavior of your code is correct. Make sure to add your tests in the corresponding file.

<!-- TODO: Add some documentation on Unit Testing for Cairo 0 hints -->