package memory

import (
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type MemoryManager struct {
	Memory *Memory
	// where the program byte code is stored
	// ProgramSegment *MemorySegment
	//// stores the function callstack (fp, ap, local variables)
	// ExecutionSegment *MemorySegment
	//// general purpose, for dynamic allocation
	// UserSegment *MemorySegment
	//// each builtin segments is stored here
	// BuiltinSegments *[]MemorySegment
}

func CreateMemoryManager(programBytecode *[]f.Element) (*MemoryManager, error) {
	memory := InitializeEmptyMemory()

	err := memory.LoadBytecode(programBytecode)
	if err != nil {
		return nil, fmt.Errorf("error creating MemoryManager: %w", err)
	}

	return &MemoryManager{
		Memory: memory,
	}, nil
}
