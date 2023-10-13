package builtins

import (
	"errors"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

type Output struct{}

func (o *Output) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (o *Output) InferValue(segment *memory.Segment, offset uint64) error {
	return errors.New("cannot infer value")
}

func (o *Output) String() string {
	return "output"
}
