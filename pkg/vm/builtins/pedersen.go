package builtins

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	pedersenhash "github.com/consensys/gnark-crypto/ecc/stark-curve/pedersen-hash"
)

const PedersenName = "pedersen"
const cellsPerPedersen = 3
const inputCellsPerPedersen = 2
const instancesPerComponentPedersen = 1

type Pedersen struct {
	ratio uint64
}

func (p *Pedersen) CheckWrite(segment *mem.Segment, offset uint64, value *mem.MemoryValue) error {
	return nil
}

func (p *Pedersen) InferValue(segment *mem.Segment, offset uint64) error {
	hashIndex := offset % cellsPerPedersen
	// input cell
	if hashIndex < inputCellsPerPedersen {
		return errors.New("cannot infer value")
	}

	xOffset := offset - hashIndex
	yOffset := xOffset + 1

	xValue := segment.Peek(xOffset)
	if !xValue.Known() {
		return fmt.Errorf("cannot infer value: input value at offset %d is unknown", xOffset)
	}

	yValue := segment.Peek(yOffset)
	if !yValue.Known() {
		return fmt.Errorf("cannot infer value: input value at offset %d is unknown", yOffset)
	}

	xFelt, err := xValue.FieldElement()
	if err != nil {
		return err
	}

	yFelt, err := yValue.FieldElement()
	if err != nil {
		return err
	}

	hash := pedersenhash.Pedersen(xFelt, yFelt)
	hashValue := mem.MemoryValueFromFieldElement(&hash)
	return segment.Write(xOffset+2, &hashValue)
}

func (p *Pedersen) String() string {
	return PedersenName
}

func (p *Pedersen) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, p.ratio, inputCellsPerPedersen, instancesPerComponentPedersen, cellsPerPedersen)
}

type AirPrivateBuiltinPedersen struct {
	Index int    `json:"index"`
	X     string `json:"x"`
	Y     string `json:"y"`
}

func (p *Pedersen) GetAirPrivateInput(pedersenSegment *mem.Segment) []AirPrivateBuiltinPedersen {
	valueMapping := make(map[int]AirPrivateBuiltinPedersen)
	for index, value := range pedersenSegment.Data {
		if !value.Known() {
			continue
		}
		idx, typ := index/cellsPerPedersen, index%cellsPerPedersen
		if typ == 2 {
			continue
		}

		builtinValue, exists := valueMapping[idx]
		if !exists {
			builtinValue = AirPrivateBuiltinPedersen{Index: idx}
		}

		valueBig := big.Int{}
		value.Felt.BigInt(&valueBig)
		valueHex := fmt.Sprintf("0x%x", &valueBig)
		if typ == 0 {
			builtinValue.X = valueHex
		} else {
			builtinValue.Y = valueHex
		}
		valueMapping[idx] = builtinValue
	}

	values := make([]AirPrivateBuiltinPedersen, 0)

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
