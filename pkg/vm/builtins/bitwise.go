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

// TODO: This is from layout small, those values should be dynamically loaded from given layout
// const ratioBitwise = 8
// const instancesPerComponentBitwise = 5

type Bitwise struct {
	ratioBitwise                 uint64
	instancesPerComponentBitwise uint64
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
	for i := 0; i < 32; i++ {
		bitwiseBytes[i] = xBytes[i] & yBytes[i]
	}
	bitwiseFelt.SetBytes(bitwiseBytes[:])
	bitwiseValue = memory.MemoryValueFromFieldElement(&bitwiseFelt)
	if err := segment.Write(xOffset+2, &bitwiseValue); err != nil {
		return err
	}

	for i := 0; i < 32; i++ {
		bitwiseBytes[i] = xBytes[i] ^ yBytes[i]
	}
	bitwiseFelt.SetBytes(bitwiseBytes[:])
	bitwiseValue = memory.MemoryValueFromFieldElement(&bitwiseFelt)
	if err := segment.Write(xOffset+3, &bitwiseValue); err != nil {
		return err
	}

	for i := 0; i < 32; i++ {
		bitwiseBytes[i] = xBytes[i] | yBytes[i]
	}
	bitwiseFelt.SetBytes(bitwiseBytes[:])
	bitwiseValue = memory.MemoryValueFromFieldElement(&bitwiseFelt)
	if err := segment.Write(xOffset+4, &bitwiseValue); err != nil {
		return err
	}

	return nil
}

func (b *Bitwise) String() string {
	return BitwiseName
}

func (b *Bitwise) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	allocatedInstances, err := GetAllocatedInstances(b.ratioBitwise, inputCellsPerBitwise, segmentUsedSize, b.instancesPerComponentBitwise, vmCurrentStep)
	if err != nil {
		return 0, err
	}
	return allocatedInstances * cellsPerBitwise, nil
}
