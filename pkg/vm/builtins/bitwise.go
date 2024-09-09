package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const BitwiseName = "bitwise"

const cellsPerBitwise = 5
const inputCellsPerBitwise = 2
const instancesPerComponentBitwise = 1

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

func (b *Bitwise) GetUsedDilutedCheckUnits(dilutedSpacing uint32, dilutedNBits uint32) uint64 {
	totalNBits := uint32(251)
	partition := make([]uint32, 0, totalNBits)

	for i := uint32(0); i < totalNBits; i += dilutedSpacing * dilutedNBits {
		for j := uint32(0); j < dilutedSpacing; j++ {
			if i+j < totalNBits {
				partition = append(partition, i+j)
			}
		}
	}

	partitionLength := uint64(len(partition))
	numTrimmed := uint64(0)

	for _, elem := range partition {
		if elem+dilutedSpacing*(dilutedNBits-1)+1 > totalNBits {
			numTrimmed++
		}
	}

	return 4*partitionLength + numTrimmed
}
