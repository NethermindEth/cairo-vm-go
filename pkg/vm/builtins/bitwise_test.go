package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBitwise(t *testing.T) {
	bitwise := &Bitwise{}
	segment := memory.EmptySegmentWithLength(5)
	segment.WithBuiltinRunner(bitwise)

	x, _ := new(fp.Element).SetString("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	y, _ := new(fp.Element).SetString("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	xValue := memory.MemoryValueFromFieldElement(x)
	yValue := memory.MemoryValueFromFieldElement(y)
	require.NoError(t, segment.Write(0, &xValue))
	require.NoError(t, segment.Write(1, &yValue))

	xAndY, err := segment.Read(2)
	require.NoError(t, err)
	xAndYFelt, err := xAndY.FieldElement()
	require.NoError(t, err)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", xAndYFelt.Text(16))

	xXorY, err := segment.Read(3)
	require.NoError(t, err)
	xXorYFelt, err := xXorY.FieldElement()
	require.NoError(t, err)
	assert.Equal(t, "11111111111111111111111111111111111111111111111111111111111111", xXorYFelt.Text(16))

	xOrY, err := segment.Read(4)
	require.NoError(t, err)
	xOrYFelt, err := xOrY.FieldElement()
	require.NoError(t, err)
	assert.Equal(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", xOrYFelt.Text(16))
}
