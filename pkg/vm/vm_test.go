package vm

import (
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVMCreation(t *testing.T) {
	const bytecodeSize = 4
	dummyBytecode := [bytecodeSize]*f.Element{
		newElementPtr(2),
		newElementPtr(3),
		newElementPtr(5),
		newElementPtr(7),
	}
	vm, err := NewVirtualMachine(dummyBytecode[:], VirtualMachineConfig{false, false})
	assert.Nil(t, err)
	assert.NotNil(t, vm)

	assert.Equal(t, 3, len(vm.MemoryManager.Memory.Segments))
	assert.Equal(t, bytecodeSize, len(vm.MemoryManager.Memory.Segments[programSegment].Data))
	assert.Equal(t, 0, len(vm.MemoryManager.Memory.Segments[executionSegment].Data))
	assert.Equal(t, 0, len(vm.MemoryManager.Memory.Segments[dataSegment].Data))
}

// todo(rodro): test all different ways of updating the ap, fp, pc
// todo(rodro): test all possible ways of that you can store values

func TestGetCellApDst(t *testing.T) {

}

func TestGetCellFpDst(t *testing.T) {

}

// create a pointer to an Element
func newElementPtr(val uint64) *f.Element {
	element := f.NewElement(val)
	return &element
}
