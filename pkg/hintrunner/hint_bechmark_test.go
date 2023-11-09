package hintrunner

import (
	"encoding/binary"
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
		lhs := Immediate(randomFeltElement(rand))

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

	rand := defaultRandGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := Immediate(randomFeltElement(rand))
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
		lhs := Immediate(randomFeltElementU128(rand))
		rhs := Immediate(randomFeltElementU128(rand))

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
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	binary.BigEndian.PutUint64(b[8:16], rand.Uint64())
	//Limit to 59 bits so at max we have a 251 bit number
	binary.BigEndian.PutUint64(b[0:8], rand.Uint64()>>5)
	f, _ := f.BigEndian.Element(&b)
	return f
}

func randomFeltElementU128(rand *rand.Rand) f.Element {
	b := [32]byte{}
	binary.BigEndian.PutUint64(b[24:32], rand.Uint64())
	binary.BigEndian.PutUint64(b[16:24], rand.Uint64())
	f, _ := f.BigEndian.Element(&b)
	return f
}

func defaultRandGenerator() *rand.Rand {
	return rand.New(rand.NewSource(0))
}

func BenchmarkUint256SquareRoot(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	rand := defaultRandGenerator()

	var sqrt0 ApCellRef = 1
	var sqrt1 ApCellRef = 2
	var remainderLow ApCellRef = 3
	var remainderHigh ApCellRef = 4
	var sqrtMul2MinusRemainderGeU128 ApCellRef = 5

	for i := 0; i < b.N; i++ {
		valueLow := Immediate(randomFeltElement(rand))
		valueHigh := Immediate(randomFeltElement(rand))
		hint := Uint256SquareRoot{
			valueLow:                     valueLow,
			valueHigh:                    valueHigh,
			sqrt0:                        sqrt0,
			sqrt1:                        sqrt1,
			remainderLow:                 remainderLow,
			remainderHigh:                remainderHigh,
			sqrtMul2MinusRemainderGeU128: sqrtMul2MinusRemainderGeU128,
		}

		err := hint.Execute(vm)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 5
	}

}
