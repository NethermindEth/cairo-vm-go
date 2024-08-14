package builtins

import (
	"errors"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const PoseidonName = "poseidon"
const CellsPerPoseidon = 6
const InputCellsPerPoseidon = 3
const instancesPerComponentPoseidon = 1

type Poseidon struct {
	ratio uint64
	cache map[uint64]fp.Element
}

func (p *Poseidon) CheckWrite(segment *mem.Segment, offset uint64, value *mem.MemoryValue) error {
	return nil
}

func (p *Poseidon) InferValue(segment *mem.Segment, offset uint64) error {
	value, ok := p.cache[offset]
	if ok {
		mv := mem.MemoryValueFromFieldElement(&value)
		return segment.Write(offset, &mv)
	}
	poseidonIndex := offset % CellsPerPoseidon
	if poseidonIndex < InputCellsPerPoseidon {
		return errors.New("cannot infer value")
	}
	baseOffset := offset - poseidonIndex
	poseidonInputValues := make([]*fp.Element, InputCellsPerPoseidon)
	for i := 0; i < InputCellsPerPoseidon; i++ {
		mv := segment.Peek(baseOffset + uint64(i))
		if !mv.Known() {
			return errors.New("cannot infer value")
		}
		poseidonInputValue, err := mv.FieldElement()
		if err != nil {
			return err
		}
		poseidonInputValues[i] = poseidonInputValue
	}

	// poseidon hash calculation
	hash := PoseidonPerm(poseidonInputValues[0], poseidonInputValues[1], poseidonInputValues[2])
	for i := 0; i < 3; i++ {
		p.cache[offset+uint64(i)] = hash[i]
	}
	value = p.cache[offset]
	mv := mem.MemoryValueFromFieldElement(&value)
	return segment.Write(offset, &mv)
}

func (p *Poseidon) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, p.ratio, InputCellsPerPoseidon, instancesPerComponentPoseidon, CellsPerPoseidon)
}

func (p *Poseidon) String() string {
	return PoseidonName
}
