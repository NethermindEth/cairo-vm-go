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

	x, _ := new(fp.Element).SetString("0x268c44203f1c763bca21beb5aec78b9063cdcdd0fdf6b598bb8e1e8f2b6253f")
	y, _ := new(fp.Element).SetString("0x2b85c9f686f5d3036db55b2ca58a763a3065bc1bc8efbe0e70f3a7171f6cad3")
	z, _ := new(fp.Element).SetString("0x61df3789eef0e1ee0dbe010582a00dd099191e6395dfb976e7be3be2fa9d54b")
	xValue := memory.MemoryValueFromFieldElement(x)
	yValue := memory.MemoryValueFromFieldElement(y)
	zValue := memory.MemoryValueFromFieldElement(z)
	require.NoError(t, segment.Write(0, &xValue))
	require.NoError(t, segment.Write(1, &yValue))
	require.NoError(t, segment.Write(2, &zValue))

	poseidonXY, err := segment.Read(5)
	require.NoError(t, err)
	pedersenXYFelt, err := poseidonXY.FieldElement()
	require.NoError(t, err)
	assert.Equal(t, "749d4d0ddf41548e039f183b745a08b80fad54e9ac389021148350bdda70a92", pedersenXYFelt.Text(16))
}
