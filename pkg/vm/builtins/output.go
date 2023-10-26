package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

const OutputName = "output"

type Output struct{}

func (o *Output) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	if !value.IsFelt() {
		return fmt.Errorf("expected a felt but got an address: %s", value)
	}
	return nil
}

func (o *Output) InferValue(segment *memory.Segment, offset uint64) error {
	return errors.New("cannot infer value")
}

func (o *Output) String() string {
	return OutputName
}
