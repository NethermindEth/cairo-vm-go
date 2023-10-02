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
	// this begins at one, because the prover expects for max memory used to
	var maxMemoryUsed uint64 = 1

	// segmentsOffsets[0] = 1
	// segmentsOffsets[1] = 1 + len(segment[0])
	// segmentsOffsets[N] = 1 + len(segment[n-1]) + sum of segements[n-1-i] for i in [1, n-1]
	segmentsOffsets := make([]uint64, uint64(len(mm.Memory.Segments))+1)
	segmentsOffsets[0] = 1
	for i, segment := range mm.Memory.Segments {
		segmentLength := segment.Len()
		maxMemoryUsed += segmentLength
		segmentsOffsets[i+1] = segmentsOffsets[i] + segmentLength
	}

	// the prover expect first element of the relocated memory to start at index 1,
	// this way we fill relocatedMemory starting from zero, but the actual value
	// returned has nil as its first element.
	relocatedMemory := make([]*f.Element, maxMemoryUsed)
	for i, segment := range mm.Memory.Segments {
		// fmt.Printf("s: %s", segment)
		for j := uint64(0); j < segment.Len(); j++ {
			cell := segment.Data[j]
			if cell == nil || !cell.Accessed {
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
