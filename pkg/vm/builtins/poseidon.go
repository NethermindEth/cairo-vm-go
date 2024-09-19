package builtins

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const PoseidonName = "poseidon"
const cellsPerPoseidon = 6
const inputCellsPerPoseidon = 3
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
		p.cache[offset+uint64(i)] = hash[i]
	}
	value = p.cache[offset]
	mv := mem.MemoryValueFromFieldElement(&value)
	return segment.Write(offset, &mv)
}

func (p *Poseidon) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, p.ratio, inputCellsPerPoseidon, instancesPerComponentPoseidon, cellsPerPoseidon)
}

func (p *Poseidon) String() string {
	return PoseidonName
}

type AirPrivateBuiltinPoseidon struct {
	Index   int    `json:"index"`
	InputS0 string `json:"input_s0"`
	InputS1 string `json:"input_s1"`
	InputS2 string `json:"input_s2"`
}

func (p *Poseidon) GetAirPrivateInput(poseidonSegment *mem.Segment) []AirPrivateBuiltinPoseidon {
	valueMapping := make(map[int]AirPrivateBuiltinPoseidon)
	for index, value := range poseidonSegment.Data {
		if !value.Known() {
			continue
		}
		idx, stateIndex := index/cellsPerPoseidon, index%cellsPerPoseidon
		if stateIndex >= inputCellsPerPoseidon {
			continue
		}

		builtinValue, exists := valueMapping[idx]
		if !exists {
			builtinValue = AirPrivateBuiltinPoseidon{Index: idx}
		}

		valueBig := big.Int{}
		value.Felt.BigInt(&valueBig)
		valueHex := fmt.Sprintf("0x%x", &valueBig)
		if stateIndex == 0 {
			builtinValue.InputS0 = valueHex
		} else if stateIndex == 1 {
			builtinValue.InputS1 = valueHex
		} else if stateIndex == 2 {
			builtinValue.InputS2 = valueHex
		}
		valueMapping[idx] = builtinValue
	}

	values := make([]AirPrivateBuiltinPoseidon, 0)

	sortedIndexes := make([]int, 0, len(valueMapping))
	for index := range valueMapping {
		sortedIndexes = append(sortedIndexes, index)
	}
	sort.Ints(sortedIndexes)
	for _, index := range sortedIndexes {
		value := valueMapping[index]
		values = append(values, value)
	}
	return values
}
