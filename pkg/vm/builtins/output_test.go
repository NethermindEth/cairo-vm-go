package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	output := &Output{}
	segment := memory.EmptySegmentWithLength(5).WithBuiltinRunner(output)

	mv1 := memory.MemoryValueFromInt(5)
	err := segment.Write(0, &mv1)
	require.NoError(t, err)

	mv2 := memory.MemoryValueFromSegmentAndOffset(1, 2)
	err = segment.Write(1, &mv2)
	require.ErrorContains(t, err, "expected a felt but got an address")

}
