package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeccakBuiltin(t *testing.T) {
	keccak := &Keccak{}
	segment := memory.EmptySegmentWithLength(9)
	segment.WithBuiltinRunner(keccak)

	v0, _ := new(fp.Element).SetString("1")
	v1, _ := new(fp.Element).SetString("2")
	v2, _ := new(fp.Element).SetString("3")
	v3, _ := new(fp.Element).SetString("4")
	v4, _ := new(fp.Element).SetString("5")
	v5, _ := new(fp.Element).SetString("6")
	v6, _ := new(fp.Element).SetString("7")
	v7, _ := new(fp.Element).SetString("8")
	v00 := memory.MemoryValueFromFieldElement(v0)
	v11 := memory.MemoryValueFromFieldElement(v1)
	v22 := memory.MemoryValueFromFieldElement(v2)
	v33 := memory.MemoryValueFromFieldElement(v3)
	v44 := memory.MemoryValueFromFieldElement(v4)
	v55 := memory.MemoryValueFromFieldElement(v5)
	v66 := memory.MemoryValueFromFieldElement(v6)
	v77 := memory.MemoryValueFromFieldElement(v7)
	require.NoError(t, segment.Write(0, &v00))
	require.NoError(t, segment.Write(1, &v11))
	require.NoError(t, segment.Write(2, &v22))
	require.NoError(t, segment.Write(3, &v33))
	require.NoError(t, segment.Write(4, &v44))
	require.NoError(t, segment.Write(5, &v55))
	require.NoError(t, segment.Write(6, &v66))
	require.NoError(t, segment.Write(7, &v77))

	k, err := segment.Read(9)
	require.NoError(t, err)
	ans, err := k.FieldElement()
	require.NoError(t, err)
	expected := fp.NewElement(0)
	_, err = expected.SetString("0x7a753f70755cbbde7882962e5969b2874c2dff11a91716ab31")
	require.NoError(t, err)
	assert.Equal(t, ans, &expected)
}
