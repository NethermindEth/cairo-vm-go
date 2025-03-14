package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

const OutputName = "output"

type Output struct {
	stopPointer uint64
	pages       []Page
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

type Page struct {
	start uint64
	size  uint64
}

func (output *Output) GetOutputPublicMemory(outputSegment memory.Segment) []memory.PublicMemoryOffset {
	publicMemory := make([]memory.PublicMemoryOffset, outputSegment.Len())

	for i := uint64(0); i < outputSegment.Len(); i++ {
		publicMemory[i] = memory.PublicMemoryOffset{
			Address: uint16(i),
			Page:    0,
		}
	}

	for _, page := range output.pages {
		for index := uint64(0); index < page.size; index++ {
			publicMemory[page.start+index].Page = uint16(page.start)
		}
	}
	return publicMemory

}
