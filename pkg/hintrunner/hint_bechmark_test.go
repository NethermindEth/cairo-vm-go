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
	v := binary.BigEndian.AppendUint64(nil, rand.Uint64())
	copy(b[24:32], v)
	v = binary.BigEndian.AppendUint64(nil, rand.Uint64())
	copy(b[16:24], v)
	v = binary.BigEndian.AppendUint64(nil, rand.Uint64())
	copy(b[8:16], v)
	//Limit to 59 bits so at max we have a 251 bit number
	v = binary.BigEndian.AppendUint64(nil, rand.Uint64()>>5)
	copy(b[0:8], v)
	f, _ := f.BigEndian.Element(&b)
	return f
}

func randomFeltElementU128(rand *rand.Rand) f.Element {
	b := [32]byte{}
	v := binary.BigEndian.AppendUint64(nil, rand.Uint64())
	copy(b[24:32], v)
	v = binary.BigEndian.AppendUint64(nil, rand.Uint64())
	copy(b[16:24], v)
	f, _ := f.BigEndian.Element(&b)
	return f
}

func defaultRandGenerator() *rand.Rand {
	return rand.New(rand.NewSource(0))
}
