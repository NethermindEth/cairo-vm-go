package zero

import (
	"fmt"
	"reflect"

	"math/big"
	"testing"

	runnerutil "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

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

func bigIntString(s string, base int) *big.Int {
	valueBig, ok := new(big.Int).SetString(s, base)
	if !ok {
		panic(fmt.Errorf("string: %v base: %d to big.Int conversion failed", s, base))
	}
	return valueBig
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
		actual := runnerutil.ReadFrom(ctx.vm, vm.ExecutionSegment, ctx.vm.Context.Ap)
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

func consecutiveVarAddrResolvedValueEquals(varName string, expectedValues []*fp.Element) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		o := ctx.operanders[varName]
		addr, err := o.GetAddress(ctx.vm)
		require.NoError(t, err)
		actualAddress, err := ctx.vm.Memory.ReadFromAddressAsAddress(&addr)
		require.NoError(t, err)
		for index, expectedValue := range expectedValues {
			expectedValueAddr := memory.MemoryAddress{SegmentIndex: actualAddress.SegmentIndex, Offset: actualAddress.Offset + uint64(index)}
			actualFelt, err := ctx.vm.Memory.ReadFromAddressAsElement(&expectedValueAddr)
			require.NoError(t, err)
			require.Equal(t, &actualFelt, expectedValue, "%s[%v] value mismatch:\nhave: %v\nwant: %v", varName, index, &actualFelt, expectedValue)
		}
	}
}

func consecutiveVarValueEquals(varName string, expectedValues []*fp.Element) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		o := ctx.operanders[varName]
		addr, err := o.GetAddress(ctx.vm)
		if err != nil {
			t.Fatal(err)
		}

		for idx := 0; idx < len(expectedValues); idx++ {
			offsetAddress, err := addr.AddOffset(int16(idx))
			if err != nil {
				t.Fatal(err)
			}

			actualFelt, err := ctx.vm.Memory.ReadFromAddressAsElement(&offsetAddress)
			if err != nil {
				t.Fatal(err)
			}

			expectedFelt := expectedValues[idx]

			if !actualFelt.Equal(expectedFelt) {
				t.Fatalf("%s value mismatch at %s:\nhave: %v\nwant: %v", varName, offsetAddress, &actualFelt, expectedFelt)
			}
		}
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
		case uint64:
			{
				valueFelt := value.(uint64)
				expectedFelt := expected.(uint64)
				if valueFelt != expectedFelt {
					t.Fatalf("%s scope value mismatch:\nhave: %d\nwant: %d", varName, value, expected)
				}
			}
		case []fp.Element:
			{
				valueArray := value.([]fp.Element)
				expectedArray := expected.([]fp.Element)
				if !reflect.DeepEqual(valueArray, expectedArray) {
					t.Fatalf("%s scope value mismatch:\nhave: %v\nwant: %v", varName, value, expected)
				}
			}
		default:
			{
				if value != expected {
					t.Fatalf("%s scope value mismatch:\nhave: %v\nwant: %v", varName, value, expected)
				}
			}
		}
	}
}

func allVarValueInScopeEquals(expectedValues map[string]any) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		for varName, expected := range expectedValues {
			varValueInScopeEquals(varName, expected)(t, ctx)
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

func varListInScopeEquals(expectedValues map[string]any) func(t *testing.T, ctx *hintTestContext) {
	return func(t *testing.T, ctx *hintTestContext) {
		for varName, expected := range expectedValues {
			varValueInScopeEquals(varName, expected)(t, ctx)
		}
	}
}
