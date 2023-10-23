package hintrunner

import (
	"testing"
)

func BenchmarkAllocSegment(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	var ap ApCellRef = 5
	for i := 0; i < b.N; i++ {
		alloc := AllocSegment{ap}
		_ = alloc.Execute(vm)
		ap += 1
	}
}
