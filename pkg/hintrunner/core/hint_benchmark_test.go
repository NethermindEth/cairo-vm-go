package core

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/stretchr/testify/assert"
)

func BenchmarkAllocSegment(b *testing.B) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	var ap hinter.ApCellRef = 1

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
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dst hinter.ApCellRef = 0
	var rhsRef hinter.ApCellRef = 1
	cell := uint64(0)

	rand := utils.DefaultRandGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.WriteTo(
			vm,
			VM.ExecutionSegment,
			vm.Context.Ap+uint64(rhsRef),
			memory.MemoryValueFromInt(rand.Int63()),
		)
		rhs := hinter.Deref{Deref: rhsRef}
		lhs := hinter.Immediate(utils.RandomFeltElement(rand))

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
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dst hinter.ApCellRef = 1

	rand := utils.DefaultRandGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := hinter.Immediate(utils.RandomFeltElement(rand))
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
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstLow hinter.ApCellRef = 0
	var dstHigh hinter.ApCellRef = 1

	rand := utils.DefaultRandGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lhs := hinter.Immediate(utils.RandomFeltElementU128(rand))
		rhs := hinter.Immediate(utils.RandomFeltElementU128(rand))

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

func BenchmarkUintDivMod(b *testing.B) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	rand := utils.DefaultRandGenerator()

	var quotient hinter.ApCellRef = 1
	var remainder hinter.ApCellRef = 2

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lhs := hinter.Immediate(utils.RandomFeltElement(rand))
		rhs := hinter.Immediate(utils.RandomFeltElement(rand))
		hint := DivMod{
			lhs:       lhs,
			rhs:       rhs,
			quotient:  quotient,
			remainder: remainder,
		}

		err := hint.Execute(vm, nil)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 5
	}
}

func BenchmarkUint256DivMod(b *testing.B) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	rand := utils.DefaultRandGenerator()

	var quotient0 hinter.ApCellRef = 1
	var quotient1 hinter.ApCellRef = 2
	var remainder0 hinter.ApCellRef = 3
	var remainder1 hinter.ApCellRef = 4

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		dividend0 := hinter.Immediate(utils.RandomFeltElement(rand))
		dividend1 := hinter.Immediate(utils.RandomFeltElement(rand))
		divisor0 := hinter.Immediate(utils.RandomFeltElement(rand))
		divisor1 := hinter.Immediate(utils.RandomFeltElement(rand))

		hint := Uint256DivMod{
			dividend0,
			dividend1,
			divisor0,
			divisor1,
			quotient0,
			quotient1,
			remainder0,
			remainder1,
		}
		err := hint.Execute(vm, nil)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 6
	}
}

func BenchmarkLinearSplit(b *testing.B) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	rand := utils.DefaultRandGenerator()

	var x hinter.ApCellRef = 0
	var y hinter.ApCellRef = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := hinter.Immediate(utils.RandomFeltElement(rand))
		scalar := hinter.Immediate(utils.RandomFeltElement(rand))
		maxX := hinter.Immediate(utils.RandomFeltElement(rand))
		hint := LinearSplit{
			value:  value,
			scalar: scalar,
			maxX:   maxX,
			x:      x,
			y:      y,
		}

		err := hint.Execute(vm, nil)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 2
	}
}

func BenchmarkUint512DivModByUint256(b *testing.B) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	rand := utils.DefaultRandGenerator()

	var quotient0 hinter.ApCellRef = 1
	var quotient1 hinter.ApCellRef = 2
	var quotient2 hinter.ApCellRef = 3
	var quotient3 hinter.ApCellRef = 4
	var remainder0 hinter.ApCellRef = 5
	var remainder1 hinter.ApCellRef = 6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dividend0 := hinter.Immediate(utils.RandomFeltElement(rand))
		dividend1 := hinter.Immediate(utils.RandomFeltElement(rand))
		dividend2 := hinter.Immediate(utils.RandomFeltElement(rand))
		dividend3 := hinter.Immediate(utils.RandomFeltElement(rand))
		divisor0 := hinter.Immediate(utils.RandomFeltElement(rand))
		divisor1 := hinter.Immediate(utils.RandomFeltElement(rand))

		hint := Uint512DivModByUint256{
			dividend0,
			dividend1,
			dividend2,
			dividend3,
			divisor0,
			divisor1,
			quotient0,
			quotient1,
			quotient2,
			quotient3,
			remainder0,
			remainder1,
		}

		err := hint.Execute(vm, nil)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 6
	}
}

func BenchmarkUint256SquareRoot(b *testing.B) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	rand := utils.DefaultRandGenerator()

	var sqrt0 hinter.ApCellRef = 1
	var sqrt1 hinter.ApCellRef = 2
	var remainderLow hinter.ApCellRef = 3
	var remainderHigh hinter.ApCellRef = 4
	var sqrtMul2MinusRemainderGeU128 hinter.ApCellRef = 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valueLow := hinter.Immediate(utils.RandomFeltElement(rand))
		valueHigh := hinter.Immediate(utils.RandomFeltElement(rand))
		hint := Uint256SquareRoot{
			valueLow:                     valueLow,
			valueHigh:                    valueHigh,
			sqrt0:                        sqrt0,
			sqrt1:                        sqrt1,
			remainderLow:                 remainderLow,
			remainderHigh:                remainderHigh,
			sqrtMul2MinusRemainderGeU128: sqrtMul2MinusRemainderGeU128,
		}

		err := hint.Execute(vm, nil)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 5
	}
}

func BenchmarkAssertLeIsFirstArcExcluded(b *testing.B) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	ctx := hinter.HintRunnerContext{
		ExcludedArc: 0,
	}

	var skipExcludeAFlag hinter.ApCellRef = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		hint := AssertLeIsFirstArcExcluded{
			SkipExcludeAFlag: skipExcludeAFlag,
		}

		err := hint.Execute(vm, &ctx)
		if err != nil {
			b.Error(err)
			break
		}

		vm.Context.Ap += 1
	}
}

func BenchmarkAssertLeIsSecondArcExcluded(b *testing.B) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	ctx := hinter.HintRunnerContext{
		ExcludedArc: 0,
	}

	var skipExcludeBMinusA hinter.ApCellRef = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		hint := AssertLeIsSecondArcExcluded{
			SkipExcludeBMinusA: skipExcludeBMinusA,
		}

		err := hint.Execute(vm, &ctx)
		if err != nil {
			b.Error(err)
			break
		}

		vm.Context.Ap += 1
	}
}

func BenchmarkAssertLeFindSmallArc(b *testing.B) {
	vm := VM.DefaultVirtualMachine()

	rand := utils.DefaultRandGenerator()
	ctx := hinter.HintRunnerContext{
		ExcludedArc: 0,
	}

	rangeCheckPtr := vm.Memory.AllocateBuiltinSegment(&builtins.RangeCheck{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// store the range check ptr at current ap
		utils.WriteTo(
			vm,
			VM.ExecutionSegment,
			vm.Context.Ap,
			memory.MemoryValueFromMemoryAddress(&rangeCheckPtr),
		)

		r1 := utils.RandomFeltElement(rand)
		r2 := utils.RandomFeltElement(rand)
		hint := AssertLeFindSmallArc{
			A:             hinter.Immediate(r1),
			B:             hinter.Immediate(r2),
			RangeCheckPtr: hinter.Deref{Deref: hinter.ApCellRef(0)},
		}

		if err := hint.Execute(vm, &ctx); err != nil &&
			!assert.ErrorContains(b, err, "check write: 2**128 <") {
			b.FailNow()
		}

		rangeCheckPtr.Offset += 4
		vm.Context.Ap += 1
	}

}

func BenchmarkRandomEcPoint(b *testing.B) {
	vm := VM.DefaultVirtualMachine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		hint := RandomEcPoint{
			x: hinter.ApCellRef(0),
			y: hinter.ApCellRef(1),
		}

		err := hint.Execute(vm)
		if err != nil {
			b.Error(err)
			break
		}

		vm.Context.Ap += 2
	}
}

func BenchmarkFieldSqrt(b *testing.B) {
	vm := VM.DefaultVirtualMachine()

	rand := utils.DefaultRandGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		hint := FieldSqrt{
			val:  hinter.Immediate(utils.RandomFeltElement(rand)),
			sqrt: hinter.ApCellRef(0),
		}

		err := hint.Execute(vm)
		if err != nil {
			b.Error(err)
			break
		}

		vm.Context.Ap += 1
	}
}
