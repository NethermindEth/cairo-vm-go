package builtins

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoseidon(t *testing.T) {
	poseidon := &Poseidon{ratio: 32, cache: make(map[uint64]fp.Element)}
	segment := memory.EmptySegmentWithLength(3)
	segment.WithBuiltinRunner(poseidon)
	inputState := []string{
		"1",
		"2",
		"3",
	}

	for i, s := range inputState {
		felt, err := new(fp.Element).SetString(s)
		if err != nil {
			panic(err)
		}
		value := memory.MemoryValueFromFieldElement(felt)
		require.NoError(t, segment.Write(uint64(i), &value))
	}
	expectedOutputStateValues := []string{
		"fa8c9b6742b6176139365833d001e30e932a9bf7456d009b1b174f36d558c5",
		"4f04deca4cb7f9f2bd16b1d25b817ca2d16fba2151e4252a2e2111cde08bfe6",
		"58dde0a2a785b395ee2dc7b60b79e9472ab826e9bb5383a8018b59772964892",
	}

	for i, v := range expectedOutputStateValues {
		hash, err := segment.Read(uint64(i + 3))
		require.NoError(t, err)
		hashValue, err := hash.FieldElement()
		require.NoError(t, err)
		fmt.Println(v, hashValue.Text(16))
		assert.Equal(t, v, hashValue.Text(16))
	}
}
