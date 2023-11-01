package hintrunner

import (
	"math/rand"
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func BenchmarkAllocSegment(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	var ap ApCellRef = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alloc := AllocSegment{ap}
		err := alloc.Execute(vm, nil)
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

	rand := defaultRandGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeTo(
			vm,
			VM.ExecutionSegment,
			vm.Context.Ap+uint64(rhsRef),
			memory.MemoryValueFromInt(rand.Int63()),
		)
		rhs := Deref{rhsRef}
		lhs := Immediate(randomIntElement(rand))

		hint := TestLessThan{
			dst: dst,
			lhs: lhs,
			rhs: rhs,
		}

		err := hint.Execute(vm, nil)
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
		value := Immediate(f.NewElement(uint64(i * i)))
		hint := SquareRoot{
			value: value,
			dst:   dst,
		}

		err := hint.Execute(vm, nil)
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

	rand := defaultRandGenerator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lhs := Immediate(randomFeltElement(rand))
		rhs := Immediate(randomFeltElement(rand))

		hint := WideMul128{
			low:  dstLow,
			high: dstHigh,
			lhs:  lhs,
			rhs:  rhs,
		}

		err := hint.Execute(vm, nil)
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

	rand := defaultRandGenerator()

	var x ApCellRef = 0
	var y ApCellRef = 1
	for i := 0; i < b.N; i++ {
		value := Immediate(randomFeltElement(rand))
		scalar := Immediate(randomFeltElement(rand))
		maxX := Immediate(randomFeltElement(rand))
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

func randomFeltElement(rand *rand.Rand) f.Element {
	data := [4]uint64{
		rand.Uint64(),
		rand.Uint64(),
		rand.Uint64(),
		rand.Uint64(),
	}
	return f.Element(data)
}

func randomIntElement(rand *rand.Rand) f.Element {
	el := rand.Int63()
	if el > 0 {
		return f.NewElement(uint64(el))
	}

	zero := f.Element{}
	sub := f.NewElement(uint64(-el))
	zero.Sub(&zero, &sub)
	return zero
}

func randomUintElement(rand *rand.Rand) f.Element {
	return f.NewElement(rand.Uint64())
}

func defaultRandGenerator() *rand.Rand {
	return rand.New(rand.NewSource(0))
}
