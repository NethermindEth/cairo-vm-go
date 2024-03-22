package zero

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	runnerutil "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

// This file defines the common testing functions.
//
// To add new tests associated with another hints file,
// look at the zerohint_math_test.go for an example.

type hintTestContext struct {
	vm            *VM.VirtualMachine
	runnerContext *hinter.HintRunnerContext

	operanders map[string]hinter.ResOperander
}

// hintTestCase describes a single zero hint test case.
//
// It can either check for an error (assert-style tests)
// or for a VM state after the execution.
//
// Both of these tests have a trivial use cases that are covered
// by helper functions. If you need to check for multiple
// memory locations and/or probe some complicated VM/runner state,
// a custom lambda can be used.
type hintTestCase struct {
	vmInit     func(vm *VM.VirtualMachine)
	ctxInit    func(ctx *hinter.HintRunnerContext)
	operanders []*hintOperander
	makeHinter func(ctx *hintTestContext) hinter.Hinter

	// Every test case should have exactly one of these functions assigned.
	// Assigning both is invalid, assigning none is not valid either.
	check    func(t *testing.T, ctx *hintTestContext)
	errCheck func(t *testing.T, ctx *hintTestContext, err error)
}

type hintOperander struct {
	Name  string
	Value any // *fp.Element or *memory.MemoryAddress
	Kind  hintOperanderKind

	// These fields are assigned automatically by the test runner.
	memoryOffset uint64
}

type builtinReference struct {
	builtin starknet.Builtin
	offset  uint64
}

// hintOperanderKind defines how the operand is going to be constructed.
// The same Value can be accessed in various ways: it could have an immediate
// value as its source, or it could be stored somewhere in the VMs memory
// (and there are several ways to address that memory as well).
type hintOperanderKind int

// These constants don't have a proper prefix to make it more pleasant to use them
// while declaring the test tables.
//
// For most of the tests, the operander kind doesn't matter that much.
// But some hints may require operanders that can successfully perform Get()
// to retrieve an address.
const (
	// [ap+offset] | Deref{ApCellRef(offset)}
	// Requires {Name, Kind=apRelative, Value}
	apRelative hintOperanderKind = iota

	// [fp+offset] | Deref{FpCellRef(offset)}
	// Requires {Name, Kind=fpRelative, Value}
	fpRelative

	// $value | Immediate($value)
	// Requires {Name, Kind=immediate, Value}
	immediate

	// A memory cell that is allocated, but not written to yet.
	// It's allowed to write to this address once.
	// Requires {Name, Kind=uninitialized}
	uninitialized

	// A value that evaluates to an address, but does not reside in memory on its own.
	// An example is a let-bound reference, like [range_check_ptr+1]
	// Requires {Name, Kind=reference, Value=*builtinRef{...}}
	reference
)

func runHinterTests(t *testing.T, tests map[string][]hintTestCase) {
	// TODO: most (all?) hinter constructors inside a single test group
	// are identical. Can we only define it once and make it reused inside
	// the entire group?

	runTest := func(t *testing.T, tc hintTestCase) {
		// Establish an invariant that only one of the check functions is present.
		if tc.check == nil && tc.errCheck == nil {
			t.Fatalf("sanity check failed: tc.check and tc.errCheck can't both be nil")
		}
		if tc.check != nil && tc.errCheck != nil {
			t.Fatalf("sanity check failed: tc.check and tc.errCheck can't be used together")
		}

		vm := VM.DefaultVirtualMachine()
		if tc.vmInit != nil {
			tc.vmInit(vm)
		}
		
		ctx := &hinter.HintRunnerContext{}
		if tc.ctxInit != nil {
			tc.ctxInit(ctx)
		}

		testCtx := &hintTestContext{
			vm:            vm,
			runnerContext: ctx,
			operanders:    make(map[string]hinter.ResOperander),
		}

		// builtinsAllocated stores fp offsets mapping for the builtin pointers.
		// If there is only one builtin allocated at [fp-4] and so on.
		// Some tests use 0 builtins and this map will be empty.
		type allocatedBuiltin struct {
			offset uint64
			addr   memory.MemoryAddress
		}
		builtinsAllocated := map[starknet.Builtin]allocatedBuiltin{}
		for _, o := range tc.operanders {
			if o.Kind != reference {
				continue
			}
			ref, ok := o.Value.(*builtinReference)
			if !ok {
				continue
			}
			if _, ok := builtinsAllocated[ref.builtin]; ok {
				continue // Already allocated
			}
			b := builtins.Runner(ref.builtin)
			addr := testCtx.vm.Memory.AllocateBuiltinSegment(b)
			builtinsAllocated[ref.builtin] = allocatedBuiltin{
				offset: vm.Context.Ap,
				addr:   addr,
			}
			runnerutil.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, memory.MemoryValueFromMemoryAddress(&addr))
			vm.Context.Ap++
		}

		// There are always a few extra values on the memory stack
		// above FP to make AP-based and FP-based addressing makes more sense.
		// These elements are identical to their index: mem[0] is 0, mem[1] is 1.
		//
		// Since these values are *below* FP, they can be considered to be arguments
		// to the function; accessed as [fp-1], etc.
		extraValues := []*fp.Element{
			feltUint64(0), // [fp-0]
			feltUint64(1), // [fp-1]
			feltUint64(2), // [fp-2]
			feltUint64(3), // [fp-3]
		}
		for _, v := range extraValues {
			runnerutil.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, memory.MemoryValueFromFieldElement(v))
			vm.Context.Ap++
		}
		// FP points *after* the last extra value.
		vm.Context.Fp = uint64(len(extraValues))

		for _, o := range tc.operanders {
			if o.Value != nil {
				switch o.Value.(type) {
				case *fp.Element, *memory.MemoryAddress, *builtinReference:
					// OK
				default:
					panic(fmt.Sprintf("unexpected operander Value type: %T", o.Value))
				}
			}

			switch o.Kind {
			case apRelative, fpRelative:
				o.memoryOffset = vm.Context.Ap
				v, err := memory.MemoryValueFromAny(o.Value)
				if err != nil {
					panic(err) // Shound never happen due to the sanity check above
				}
				runnerutil.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, v)
				vm.Context.Ap++

			case immediate, reference:
				// Nothing to do.

			case uninitialized:
				o.memoryOffset = vm.Context.Ap
				vm.Context.Ap++

			default:
				panic("unexpected operander kind")
			}
		}

		// Now that we filled the memory with values, we can
		// compute the relative addresses for operanders.
		for _, o := range tc.operanders {
			switch o.Kind {
			case apRelative, uninitialized:
				relOffset := int(vm.Context.Ap - o.memoryOffset)
				testCtx.operanders[o.Name] = &hinter.Deref{
					Deref: hinter.ApCellRef(-relOffset),
				}

			case fpRelative:
				relOffset := int(vm.Context.Fp - o.memoryOffset)
				testCtx.operanders[o.Name] = &hinter.Deref{
					Deref: hinter.FpCellRef(-relOffset),
				}

			case immediate:
				testCtx.operanders[o.Name] = hinter.Immediate(*o.Value.(*fp.Element))

			case reference:
				switch value := o.Value.(type) {
				case *builtinReference:
					// value.offset is an offset relative to the builtin pointer,
					// builtin.offset is an offset that locates the builtin pointer inside the memory (fp-relative).
					// Therefore, [[fp+relOffset] + value.offset] produces the final address.
					builtin := builtinsAllocated[value.builtin]
					relOffset := int(vm.Context.Fp + builtin.offset)
					testCtx.operanders[o.Name] = &hinter.DoubleDeref{
						Deref: hinter.Deref{
							Deref: hinter.FpCellRef(-relOffset),
						},
						Offset: int16(value.offset),
					}
				default:
					panic(fmt.Sprintf("unsupported reference type: %T", value))
				}
			}
		}

		h := tc.makeHinter(testCtx)

		err := h.Execute(vm, ctx)

		if tc.errCheck != nil {
			// Error checking test.
			tc.errCheck(t, testCtx, err)
			return
		}

		// VM state checking test.
		require.Nil(t, err)
		tc.check(t, testCtx)
	}

	for testGroup, cases := range tests {
		{
			// A sanity check: test that there are no duplicated test cases inside a group.
			type testCaseKey struct {
				Operanders   []*hintOperander
				IsErrorCheck bool
			}
			set := map[string]struct{}{}
			for i, tc := range cases {
				key := testCaseKey{
					Operanders:   tc.operanders,
					IsErrorCheck: tc.errCheck != nil,
				}
				stringKey, err := json.Marshal(key)
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := set[string(stringKey)]; ok {
					t.Fatalf("%s: duplicated test case case (i=%d) found: %s", testGroup, i, stringKey)
				}
				set[string(stringKey)] = struct{}{}
			}
		}

		for i, tc := range cases {
			t.Run(fmt.Sprintf("%s_%d", testGroup, i), func(t *testing.T) {
				runTest(t, tc)
			})
		}
	}
}
