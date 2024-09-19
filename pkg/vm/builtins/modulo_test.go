package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func TestModuloBuiltin(t *testing.T) {
	mod := &ModBuiltin{ratio: 2048, modBuiltinType: Add}
	segment := memory.EmptySegmentWithLength(9)
	segment.WithBuiltinRunner(mod)
}
