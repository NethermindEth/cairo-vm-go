package builtins

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const (
	BitwiseName                  = "bitwise"
	cellsPerBitwise              = 5
	inputCellsPerBitwise         = 2
	instancesPerComponentBitwise = 1
)

type Bitwise struct {
	ratio uint64
}

func (b *Bitwise) CheckWrite(
	segment *memory.Segment, offset uint64, value *memory.MemoryValue,
) error {
	return nil
}

func (b *Bitwise) InferValue(segment *memory.Segment, offset uint64) error {
	bitwiseIndex := offset % cellsPerBitwise
	// input cell
	if bitwiseIndex < inputCellsPerBitwise {
		return errors.New("cannot infer value from input cell")
	}

	xOffset := offset - bitwiseIndex
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

	xBytes := xFelt.Bytes()
	yBytes := yFelt.Bytes()

	var bitwiseValue memory.MemoryValue
	var bitwiseFelt fp.Element
	var bitwiseBytes [32]byte

	if bitwiseIndex == 2 {
		for i := 0; i < 32; i++ {
			bitwiseBytes[i] = xBytes[i] & yBytes[i]
		}
		bitwiseFelt.SetBytes(bitwiseBytes[:])
		bitwiseValue = memory.MemoryValueFromFieldElement(&bitwiseFelt)
		if err := segment.Write(xOffset+2, &bitwiseValue); err != nil {
			return err
		}
	} else if bitwiseIndex == 3 {
		for i := 0; i < 32; i++ {
			bitwiseBytes[i] = xBytes[i] ^ yBytes[i]
		}
		bitwiseFelt.SetBytes(bitwiseBytes[:])
		bitwiseValue = memory.MemoryValueFromFieldElement(&bitwiseFelt)
		if err := segment.Write(xOffset+3, &bitwiseValue); err != nil {
			return err
		}
	} else if bitwiseIndex == 4 {
		for i := 0; i < 32; i++ {
			bitwiseBytes[i] = xBytes[i] | yBytes[i]
		}
		bitwiseFelt.SetBytes(bitwiseBytes[:])
		bitwiseValue = memory.MemoryValueFromFieldElement(&bitwiseFelt)
		if err := segment.Write(xOffset+4, &bitwiseValue); err != nil {
			return err
		}
	}

	return nil
}

func (b *Bitwise) String() string {
	return BitwiseName
}

func (b *Bitwise) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, b.ratio, inputCellsPerBitwise, instancesPerComponentBitwise, cellsPerBitwise)
}

type AirPrivateBuiltinBitwise struct {
	Index int    `json:"index"`
	X     string `json:"x"`
	Y     string `json:"y"`
}

func (b *Bitwise) GetAirPrivateInput(bitwiseSegment *memory.Segment) []AirPrivateBuiltinBitwise {
	valueMapping := make(map[int]AirPrivateBuiltinBitwise)
	for index, value := range bitwiseSegment.Data {
		if !value.Known() {
			continue
		}
		idx, typ := index/cellsPerBitwise, index%cellsPerBitwise
		if typ >= 2 {
			continue
		}

		builtinValue, exists := valueMapping[idx]
		if !exists {
			builtinValue = AirPrivateBuiltinBitwise{Index: idx}
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

	values := make([]AirPrivateBuiltinBitwise, 0)

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
