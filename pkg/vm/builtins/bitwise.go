package builtins

import (
	"errors"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const cellsPerBitwise = 5
const inputCellsPerBitwise = 2

type Bitwise struct{}

func (b *Bitwise) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (b *Bitwise) InferValue(segment *memory.Segment, offset uint64) error {
	bitwiseIndex := offset % cellsPerBitwise
	// input cell
	if bitwiseIndex < inputCellsPerBitwise {
		return errors.New("cannot infer value")
	}

	xOffset := offset - bitwiseIndex
	yOffset := xOffset + 1

	xValue, err := segment.Read(xOffset)
	if err != nil {
		return err
	}

	yValue, err := segment.Read(yOffset)
	if err != nil {
		return err
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
	return "bitwise"
}
