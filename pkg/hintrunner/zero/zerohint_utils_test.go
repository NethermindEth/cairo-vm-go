package zero

import (
	"math/big"
	"testing"

	runnerutil "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func bigIntString(s string) *big.Int {
	i, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("failed to parse big.Int")
	}
	return i

}

func addr(offset uint64) *memory.MemoryAddress {
	return &memory.MemoryAddress{
		SegmentIndex: vm.ExecutionSegment,
		Offset:       offset,
	}
}

func addrWithSegment(segment, offset uint64) *memory.MemoryAddress {
	return &memory.MemoryAddress{
		SegmentIndex: segment,
		Offset:       offset,
	}
}

func addrBuiltin(builtin starknet.Builtin, offset uint64) *builtinReference {
	return &builtinReference{
		builtin: builtin,
		offset:  offset,
	}
}

func feltString(s string) *fp.Element {
	felt, err := new(fp.Element).SetString(s)
	if err != nil {
		panic(err)
	}
	return felt
}

func feltInt64(v int64) *fp.Element {
	return new(fp.Element).SetInt64(v)
}

func feltUint64(v uint64) *fp.Element {
	return new(fp.Element).SetUint64(v)
}

func feltAdd(x, y *fp.Element) *fp.Element {
	return new(fp.Element).Add(x, y)
}

func apValueEquals(expected *fp.Element) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		actual := runnerutil.ReadFrom(ctx.vm, VM.ExecutionSegment, ctx.vm.Context.Ap)
		actualFelt, err := actual.FieldElement()
		if err != nil {
			t.Fatal(err)
		}
		if expected.Cmp(actualFelt) != 0 {
			t.Fatalf("ap values mismatch:\nhave: %v\nwant: %v", actualFelt, expected)
		}
	}
}

func varValueEquals(varName string, expected *fp.Element) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		o := ctx.operanders[varName]
		addr, err := o.GetAddress(ctx.vm)
		if err != nil {
			t.Fatal(err)
		}
		actualFelt, err := ctx.vm.Memory.ReadFromAddressAsElement(&addr)
		if err != nil {
			t.Fatal(err)
		}
		if !actualFelt.Equal(expected) {
			t.Fatalf("%s value mismatch:\nhave: %v\nwant: %v", varName, &actualFelt, expected)
		}
	}
}

func allVarValueEquals(expectedValues map[string]*fp.Element) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		for varName, expected := range expectedValues {
			varValueEquals(varName, expected)(t, ctx)
		}
	}
}

func errorTextContains(s string) func(t *testing.T, ctx *hintTestContext, err error) {
	return func(t *testing.T, ctx *hintTestContext, err error) {
		if err == nil {
			t.Fatalf("expected an error containing %q, got nil err", s)
		}
		require.ErrorContains(t, err, s)
	}
}

func errorIsNil(t *testing.T, ctx *hintTestContext, err error) {
	if err != nil {
		t.Fatalf("expected a nil error, got: %v", err)
	}
}

func varValueInScopeEquals(varName string, expected any) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		value, err := ctx.runnerContext.ScopeManager.GetVariableValue(varName)
		if err != nil {
			t.Fatal(err)
		}
		switch expected.(type) {
		case *big.Int:
			{
				valueBig := value.(*big.Int)
				expectedBig := expected.(*big.Int)
				if valueBig.Cmp(expectedBig) != 0 {
					t.Fatalf("%s scope value mismatch:\nhave: %v\nwant: %v", varName, value, expected)
				}
			}
		case *fp.Element:
			{
				valueFelt := value.(*fp.Element)
				expectedFelt := expected.(*fp.Element)
				if valueFelt.Cmp(expectedFelt) != 0 {
					t.Fatalf("%s scope value mismatch:\nhave: %v\nwant: %v", varName, value, expected)
				}
			}
		}
	}
}

func varListInScopeEquals(expectedValues map[string]any) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		for varName, expected := range expectedValues {
			varValueInScopeEquals(varName, expected)(t, ctx)
		}
	}
}
