package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

const OutputName = "output"

type Output struct {
	stopPointer uint64
}

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

func (o *Output) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return segmentUsedSize, nil
}

func (o *Output) GetCellsPerInstance() uint64 {
	return 0
}

func (o *Output) GetStopPointer() uint64 {
	return o.stopPointer
}

func (o *Output) SetStopPointer(stopPointer uint64) {
	o.stopPointer = stopPointer
}
