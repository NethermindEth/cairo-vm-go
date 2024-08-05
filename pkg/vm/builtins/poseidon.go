package builtins

import (
	"errors"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const PoseidonName = "poseidon"
const cellsPerPoseidon = 6
const inputCellsPerPoseidon = 3
const instancesPerComponentPoseidon = 1

type Poseidon struct {
	ratio uint64
	cache map[uint64]mem.MemoryValue
}

func (p *Poseidon) CheckWrite(segment *mem.Segment, offset uint64, value *mem.MemoryValue) error {
	return nil
}

func (p *Poseidon) InferValue(segment *mem.Segment, offset uint64) error {
	val, ok := p.cache[offset]
	if ok {
		return segment.Write(offset, &val)
	}
	poseidonIndex := offset % cellsPerPoseidon
	if poseidonIndex < inputCellsPerPoseidon {
		return errors.New("cannot infer value")
	}
	baseOffset := offset - poseidonIndex
	poseidonInputValues := make([]*fp.Element, inputCellsPerPoseidon)
	for i := 0; i < inputCellsPerPoseidon; i++ {
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
		hashValue := mem.MemoryValueFromFieldElement(&hash[i])
		p.cache[offset+uint64(i)] = hashValue
	}
	value := p.cache[offset]
	return segment.Write(offset, &value)
}

func (p *Poseidon) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, p.ratio, inputCellsPerPoseidon, instancesPerComponentPoseidon, cellsPerPoseidon)
}

func (p *Poseidon) String() string {
	return PoseidonName
}
