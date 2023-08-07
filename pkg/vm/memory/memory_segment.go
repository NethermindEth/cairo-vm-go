package memory

import (
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type MemoryManager struct {
	Memory Memory
	// where the program byte code is stored
	// ProgramSegment *MemorySegment
	//// stores the function callstack (fp, ap, local variables)
	// ExecutionSegment *MemorySegment
	//// general purpose, for dynamic allocation
	// UserSegment *MemorySegment
	//// each builtin segments is stored here
	// BuiltinSegments *[]MemorySegment
}

func CreateMemoryManager(programBytecode *[]f.Element) MemoryManager {
	// memory := InitializeEmptyMemory()

	return MemoryManager{
		//ProgramSegment:   &MemorySegment{Cells: *programBytecode},
		//ExecutionSegment: nil,
		//UserSegment:      nil,
		//BuiltinSegments:  nil,
	}
}
