package memory

import (
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type MemoryManager struct {
	Memory *Memory
}

// Creates a new memory manager
func CreateMemoryManager() (*MemoryManager, error) {
	memory := InitializeEmptyMemory()

	return &MemoryManager{
		Memory: memory,
	}, nil
}

func (mm *MemoryManager) RelocateMemory() []*f.Element {
	maxMemoryUsed := 0
	// segmentsOffsets[i] =  sum of segmentsOffset[i - 1] + [i - 2] ... [0]
	segmentsOffsets := make([]int, len(mm.Memory.Segments))
	for i, segment := range mm.Memory.Segments {
		maxMemoryUsed += len(segment.Data)
		if i == 0 {
			segmentsOffsets[i] = 0
		} else {
			segmentsOffsets[i] = segmentsOffsets[i-1] + len(segment.Data)
		}
	}

	relocatedMemory := make([]*f.Element, maxMemoryUsed)
	for i, segment := range mm.Memory.Segments {
		for j, cell := range segment.Data {
			if !cell.Accessed {
				continue
			}
			var felt *f.Element
			if cell.Value.IsAddress() {
				felt = cell.Value.address.Relocate(segmentsOffsets)
			} else {
				felt = cell.Value.felt
			}

			relocatedMemory[segmentsOffsets[i]+j] = felt
		}
	}

	return relocatedMemory
}
