package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoseidon(t *testing.T) {
	poseidon := &Poseidon{}
	segment := memory.EmptySegmentWithLength(3)
	segment.WithBuiltinRunner(poseidon)

	x, _ := new(fp.Element).SetString("0xb662f9017fa7956fd70e26129b1833e10ad000fd37b4d9f4e0ce6884b7bbe")
	y, _ := new(fp.Element).SetString("0x1fe356bf76102cdae1bfbdc173602ead228b12904c00dad9cf16e035468bea")
	xValue := memory.MemoryValueFromFieldElement(x)
	yValue := memory.MemoryValueFromFieldElement(y)
	require.NoError(t, segment.Write(0, &xValue))
	require.NoError(t, segment.Write(1, &yValue))

	poseidonXY, err := segment.Read(2)
	require.NoError(t, err)
	poseidonXYFelt, err := poseidonXY.FieldElement()
	require.NoError(t, err)
	assert.Equal(t, "75540825a6ecc5dc7d7c2f5f868164182742227f1367d66c43ee51ec7937a81", poseidonXYFelt.Text(16))
}
