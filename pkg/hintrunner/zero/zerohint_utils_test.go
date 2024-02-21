package zero

import (
	"testing"

	runnerutil "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

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
