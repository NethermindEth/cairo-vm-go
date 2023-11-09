package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

  junoFelt "github.com/NethermindEth/juno/core/felt"
  junoCrypto "github.com/NethermindEth/juno/core/crypto"
)

const PoseidonName = "pedersen"
const cellsPerPoseidon = 3
const inputCellsPerPoseidon = 2

type Poseidon struct{}

func (p *Poseidon) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (p *Poseidon) InferValue(segment *memory.Segment, offset uint64) error {
	hashIndex := offset % cellsPerPoseidon
	// input cell
	if hashIndex < inputCellsPerPoseidon {
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
  
  junoX, junoY := junoFelt.NewFelt(xFelt), junoFelt.NewFelt(yFelt) 
	hash := junoCrypto.Poseidon(junoX, junoY)
	hashValue := memory.MemoryValueFromFieldElement(hash.Impl())
	return segment.Write(xOffset+2, &hashValue)
}

func (p *Poseidon) String() string {
	return PoseidonName
}
