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
}

func (p *Poseidon) CheckWrite(segment *mem.Segment, offset uint64, value *mem.MemoryValue) error {
	return nil
}

func (p *Poseidon) InferValue(segment *mem.Segment, offset uint64) error {
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
		err := segment.Write(baseOffset+uint64(i+3), &hashValue)
		if err != nil {
			return err
		}

	}
	return nil
}

func (p *Poseidon) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, p.ratio, inputCellsPerPoseidon, instancesPerComponentPoseidon, cellsPerPoseidon)
}

func (p *Poseidon) String() string {
	return PoseidonName
}
