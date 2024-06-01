package builtins

import (
	"errors"
	"fmt"

	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	pedersenhash "github.com/consensys/gnark-crypto/ecc/stark-curve/pedersen-hash"
)

const PedersenName = "pedersen"
const cellsPerPedersen = 3
const inputCellsPerPedersen = 2

type Pedersen struct {
	ratioPedersen                 uint64
	instancesPerComponentPedersen uint64
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
	allocatedInstances, err := GetAllocatedInstances(p.ratioPedersen, inputCellsPerPedersen, segmentUsedSize, p.instancesPerComponentPedersen, vmCurrentStep)
	if err != nil {
		return 0, err
	}
	return allocatedInstances * cellsPerPedersen, nil
}
