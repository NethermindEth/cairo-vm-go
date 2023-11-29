package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPedersen(t *testing.T) {
	pedersen := &Pedersen{}
	segment := memory.EmptySegmentWithLength(3)
	segment.WithBuiltinRunner(pedersen)

	x, _ := new(fp.Element).SetString("0x03d937c035c878245caf64531a5756109c53068da139362728feb561405371cb")
	y, _ := new(fp.Element).SetString("0x0208a0a10250e382e1e4bbe2880906c2791bf6275695e02fbbc6aeff9cd8b31a")
	xValue := memory.MemoryValueFromFieldElement(x)
	yValue := memory.MemoryValueFromFieldElement(y)
	require.NoError(t, segment.Write(0, &xValue))
	require.NoError(t, segment.Write(1, &yValue))

	pedersenXY, err := segment.Read(2)
	require.NoError(t, err)
	pedersenXYFelt, err := pedersenXY.FieldElement()
	require.NoError(t, err)
	assert.Equal(t, "30e480bed5fe53fa909cc0f8c4d99b8f9f2c016be4c41e13a4848797979c662", pedersenXYFelt.Text(16))
}
