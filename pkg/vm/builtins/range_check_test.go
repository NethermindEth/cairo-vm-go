package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangeCheckWriteMemoryAddress(t *testing.T) {
	builtin := RangeCheck{0, 8}
	memoryAddress := memory.EmptyMemoryValueAsAddress()
	assert.Error(t, builtin.CheckWrite(nil, 0, &memoryAddress))
}

func TestRangeCheckWriteOutOfRange(t *testing.T) {
	builtin := RangeCheck{0, 8}
	outOfRangeValueFelt, err := new(fp.Element).SetString("0x100000000000000000000000000000001")
	require.NoError(t, err)
	outOfRangeValue := memory.MemoryValueFromFieldElement(outOfRangeValueFelt)
	assert.Error(t, builtin.CheckWrite(nil, 0, &outOfRangeValue))
}

func TestRangeCheckWrite(t *testing.T) {
	builtin := RangeCheck{0, 8}
	f, err := new(fp.Element).SetString("0x44")
	require.NoError(t, err)
	v := memory.MemoryValueFromFieldElement(f)
	assert.NoError(t, builtin.CheckWrite(nil, 0, &v))
}

func TestRangeCheckInfer(t *testing.T) {
	builtin := RangeCheck{0, 8}
	segment := memory.EmptySegmentWithLength(3)
	assert.ErrorContains(t, builtin.InferValue(segment, 0), "cannot infer value")
}

func TestRangeCheck96WriteMemoryAddress(t *testing.T) {
	builtin := RangeCheck{0, 6}
	memoryAddress := memory.EmptyMemoryValueAsAddress()
	assert.Error(t, builtin.CheckWrite(nil, 0, &memoryAddress))
}

func TestRangeCheck96WriteOutOfRange(t *testing.T) {
	builtin := RangeCheck{0, 6}
	outOfRangeValueFelt, err := new(fp.Element).SetString("40564819207303340847894502572032")
	require.NoError(t, err)
	outOfRangeValue := memory.MemoryValueFromFieldElement(outOfRangeValueFelt)
	assert.Error(t, builtin.CheckWrite(nil, 0, &outOfRangeValue))
}

func TestRangeCheck96Write(t *testing.T) {
	builtin := RangeCheck{0, 6}
	f, err := new(fp.Element).SetString("19342813113834066795298816")
	require.NoError(t, err)
	v := memory.MemoryValueFromFieldElement(f)
	assert.NoError(t, builtin.CheckWrite(nil, 0, &v))
}

func TestRangeCheck96Infer(t *testing.T) {
	builtin := RangeCheck{0, 6}
	segment := memory.EmptySegmentWithLength(3)
	assert.ErrorContains(t, builtin.InferValue(segment, 0), "cannot infer value")
}
