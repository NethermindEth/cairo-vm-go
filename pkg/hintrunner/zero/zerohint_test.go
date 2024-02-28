package zero

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	runnerutil "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
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
	vminit     func(vm *VM.VirtualMachine)
	operanders []*hintOperander
	makeHinter func(ctx *hintTestContext) hinter.Hinter

	// Every test case should have exactly one of these functions assigned.
	// Assigning both is invalid, assigning none is not valid either.
	check    func(t *testing.T, ctx *hintTestContext)
	errCheck func(t *testing.T, ctx *hintTestContext, err error)
}

type hintOperander struct {
	Name  string
	Value []*fp.Element
	Kind  hintOperanderKind

	// These fields are assigned automatically by the test runner.
	memoryOffset uint64
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
		ctx := &hinter.HintRunnerContext{}
		if tc.vminit != nil {
			tc.vminit(vm)
		}

		testCtx := &hintTestContext{
			vm:            vm,
			runnerContext: ctx,
			operanders:    make(map[string]hinter.ResOperander),
		}

		// There are always a few extra values on the memory stack
		// above FP to make AP-based and FP-based addressing makes more sense.
		// These elements are identical to their index: mem[0] is 0, mem[1] is 1.
		//
		// Since these values are *below* FP, they can be considered to be arguments
		// to the function; accessed as [fp-1], etc.
		extraValues := []*fp.Element{
			feltUint64(0),
			feltUint64(1),
			feltUint64(2),
			feltUint64(3),
		}
		for _, v := range extraValues {
			runnerutil.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, memory.MemoryValueFromFieldElement(v))
			vm.Context.Ap++
		}
		// FP points *after* the last extra value.
		vm.Context.Fp = uint64(len(extraValues))

		for _, o := range tc.operanders {
			switch o.Kind {
			case apRelative, fpRelative:
				o.memoryOffset = vm.Context.Ap
				for _, v := range o.Value {
					runnerutil.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, memory.MemoryValueFromFieldElement(v))
					vm.Context.Ap++
				}

			case immediate:
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
				testCtx.operanders[o.Name] = hinter.Immediate(*o.Value[0])
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
		for i, tc := range cases {
			t.Run(fmt.Sprintf("%s_%d", testGroup, i), func(t *testing.T) {
				runTest(t, tc)
			})
		}
	}
}
