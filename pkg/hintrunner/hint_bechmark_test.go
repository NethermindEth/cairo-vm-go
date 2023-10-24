package hintrunner

import (
	"math/big"
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
)

func BenchmarkAllocSegment(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	var ap ApCellRef = 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alloc := AllocSegment{ap}
		err := alloc.Execute(vm)
		if err != nil {
			panic(err)
		}

		vm.Context.Ap += 1
	}
}

func BenchmarkLessThan(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dst ApCellRef = 1
	var rhsRef ApCellRef = 2
	cell := uint64(0)
	for i := 0; i < b.N; i++ {
		writeTo(vm, VM.ExecutionSegment, vm.Context.Ap+uint64(rhsRef), memory.MemoryValueFromInt(int64(23*cell)%13))
		rhs := Deref{rhsRef}
		lhs := Immediate(*big.NewInt(int64((13 * cell) % 11)))
		hint := TestLessThan{
			dst: dst,
			lhs: lhs,
			rhs: rhs,
		}
		err := hint.Execute(vm)
		if err != nil {
			panic(err)
		}

		vm.Context.Ap += 3
		cell += 1
	}

}
