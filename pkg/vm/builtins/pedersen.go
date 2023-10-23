package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	pedersenhash "github.com/consensys/gnark-crypto/ecc/stark-curve/pedersen-hash"
)

const PedersenName = "pedersen"
const cellsPerPedersen = 3
const inputCellsPerPedersen = 2

type Pedersen struct{}

func (p *Pedersen) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (p *Pedersen) InferValue(segment *memory.Segment, offset uint64) error {
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
	hashValue := memory.MemoryValueFromFieldElement(&hash)
	return segment.Write(xOffset+2, &hashValue)
}

func (p *Pedersen) String() string {
	return PedersenName
}
