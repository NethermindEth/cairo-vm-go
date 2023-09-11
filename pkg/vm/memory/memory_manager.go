package memory

import (
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type MemoryManager struct {
	Memory *Memory
}

// Creates a new memory manager
func CreateMemoryManager() *MemoryManager {
	memory := InitializeEmptyMemory()

	return &MemoryManager{
		Memory: memory,
	}
}

// It returns all segments in memory but relocated as a single segment
// Each element is a pointer to a field element, if the cell was not accessed,
// nil is stored instead
func (mm *MemoryManager) RelocateMemory() []*f.Element {
	maxMemoryUsed := 0
	// segmentsOffsets[0] =  0
	// segmentsOffsets[1] = len(segment[0])
	// segmentsOffsets[N] = len(segment[n - 1]) + sum of segmentsOffsets[n - i] for i in [0, n-1]
	segmentsOffsets := make([]uint64, uint64(len(mm.Memory.Segments))+1)
	for i, segment := range mm.Memory.Segments {
		maxMemoryUsed += len(segment.Data)
		segmentsOffsets[i+1] = segmentsOffsets[i] + uint64(len(segment.Data))
	}

	relocatedMemory := make([]*f.Element, maxMemoryUsed)
	for i, segment := range mm.Memory.Segments {
		for j, cell := range segment.Data {
			var felt *f.Element
			if !cell.Accessed {
				continue
			}
			if cell.Value.IsAddress() {
				felt = cell.Value.address.Relocate(segmentsOffsets)
			} else {
				felt = cell.Value.felt
			}

			relocatedMemory[segmentsOffsets[i]+uint64(j)] = felt
		}
	}

	return relocatedMemory
}
