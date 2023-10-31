package hintrunner

import (
	"math/big"
	"math/rand"
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
			b.Error(err)
			break
		}

		vm.Context.Ap += 1
	}
}

func BenchmarkLessThan(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dst ApCellRef = 0
	var rhsRef ApCellRef = 1
	cell := uint64(0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		writeTo(vm, VM.ExecutionSegment, vm.Context.Ap+uint64(rhsRef), memory.MemoryValueFromInt(rand.Int63()))
		rhs := Deref{rhsRef}
		lhs := Immediate(*big.NewInt(rand.Int63()))

		hint := TestLessThan{
			dst: dst,
			lhs: lhs,
			rhs: rhs,
		}

		err := hint.Execute(vm)
		if err != nil {
			b.Error(err)
			break
		}

		vm.Context.Ap += 2
		cell += 1
	}
}

func BenchmarkSquareRoot(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dst ApCellRef = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//TODO: Change to rand.Uint64()
		value := Immediate(*big.NewInt(int64(i * i)))
		hint := SquareRoot{
			value: value,
			dst:   dst,
		}

		err := hint.Execute(vm)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 1

	}

}

func BenchmarkWideMul128(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstLow ApCellRef = 0
	var dstHigh ApCellRef = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lhs := Immediate(*new(big.Int).SetUint64(rand.Uint64()))
		rhs := Immediate(*new(big.Int).SetUint64(rand.Uint64()))

		hint := WideMul128{
			low:  dstLow,
			high: dstHigh,
			lhs:  lhs,
			rhs:  rhs,
		}

		err := hint.Execute(vm)
		if err != nil {
			b.Error(err)
			break
		}

		vm.Context.Ap += 2
	}
}

func BenchmarkLinearSplit(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var x ApCellRef = 0
	var y ApCellRef = 1
	for i := 0; i < b.N; i++ {
		value := Immediate(*big.NewInt(rand.Int63()))
		scalar := Immediate(*big.NewInt(rand.Int63()))
		maxX := Immediate(*big.NewInt(rand.Int63()))
		hint := LinearSplit{
			value:  value,
			scalar: scalar,
			maxX:   maxX,
			x:      x,
			y:      y,
		}

		err := hint.Execute(vm)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 2
	}

}
