package memory

import (
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type MemoryManager struct {
	Memory *Memory
}

func CreateMemoryManager() (*MemoryManager, error) {
	memory := InitializeEmptyMemory()

	return &MemoryManager{
		Memory: memory,
	}, nil
}

func (mm *MemoryManager) GetByteCodeAt(segmentIndex uint64, offset uint64) *f.Element {
	return nil
}
