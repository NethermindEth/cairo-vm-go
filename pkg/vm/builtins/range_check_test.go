package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangeCheckWriteMemoryAddress(t *testing.T) {
	builtin := RangeCheck{}
	memoryAddress := memory.EmptyMemoryValueAsAddress()
	assert.Error(t, builtin.CheckWrite(nil, 0, &memoryAddress))
}

func TestRangeCheckWriteOutOfRange(t *testing.T) {
	builtin := RangeCheck{}
	outOfRangeValueFelt, err := new(fp.Element).SetString("0x100000000000000000000000000000001")
	require.NoError(t, err)
	outOfRangeValue := memory.MemoryValueFromFieldElement(outOfRangeValueFelt)
	assert.Error(t, builtin.CheckWrite(nil, 0, &outOfRangeValue))
}

func TestRangeCheckWrite(t *testing.T) {
	builtin := RangeCheck{}
	f, err := new(fp.Element).SetString("0x44")
	require.NoError(t, err)
	v := memory.MemoryValueFromFieldElement(f)
	assert.NoError(t, builtin.CheckWrite(nil, 0, &v))
}

func TestRangeCheckInfer(t *testing.T) {
	builtin := RangeCheck{}
	segment := memory.EmptySegmentWithLength(3)
	assert.ErrorContains(t, builtin.InferValue(segment, 0), "cannot infer value")
}
