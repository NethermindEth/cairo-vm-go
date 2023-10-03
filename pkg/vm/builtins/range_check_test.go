package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangeCheck(t *testing.T) {
	builtin := RangeCheck{}

	t.Run("write memory addr", func(t *testing.T) {
		memoryAddress := memory.EmptyMemoryValueAsAddress()
		assert.Error(t, builtin.CheckWrite(nil, 0, &memoryAddress))
	})

	t.Run("write out of range", func(t *testing.T) {
		outOfRangeValueFelt, err := new(fp.Element).SetString("0x100000000000000000000000000000001")
		require.NoError(t, err)
		outOfRangeValue := memory.MemoryValueFromFieldElement(outOfRangeValueFelt)
		assert.Error(t, builtin.CheckWrite(nil, 0, &outOfRangeValue))
	})

	t.Run("write in range", func(t *testing.T) {
		f, err := new(fp.Element).SetString("0x44")
		require.NoError(t, err)
		v := memory.MemoryValueFromFieldElement(f)
		assert.NoError(t, builtin.CheckWrite(nil, 0, &v))
	})

	t.Run("deduce", func(t *testing.T) {
		segment := memory.EmptySegmentWithLength(3)
		assert.NoError(t, builtin.InferValue(segment, 0))
		require.Equal(t, memory.EmptyMemoryValueAsFelt(), segment.Data[0])
	})
}
