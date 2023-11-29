package hintrunner

import (
	"testing"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/stretchr/testify/assert"
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

		err := hint.Execute(vm, nil)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 2
	}
}

func BenchmarkUint512DivModByUint256(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	rand := defaultRandGenerator()

	var quotient0 ApCellRef = 1
	var quotient1 ApCellRef = 2
	var quotient2 ApCellRef = 3
	var quotient3 ApCellRef = 4
	var remainder0 ApCellRef = 5
	var remainder1 ApCellRef = 6

	for i := 0; i < b.N; i++ {
		dividend0 := Immediate(randomFeltElement(rand))
		dividend1 := Immediate(randomFeltElement(rand))
		dividend2 := Immediate(randomFeltElement(rand))
		dividend3 := Immediate(randomFeltElement(rand))
		divisor0 := Immediate(randomFeltElement(rand))
		divisor1 := Immediate(randomFeltElement(rand))

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

		err := hint.Execute(vm, nil)
		if err != nil {
			b.Error(err)
			break
		}
		vm.Context.Ap += 5
	}
}

func BenchmarkAssertLeIsFirstArcExcluded(b *testing.B) {
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	ctx := HintRunnerContext{
		ExcludedArc: 0,
	}

	var skipExcludeAFlag ApCellRef = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		hint := AssertLeIsFirstArcExcluded{
			skipExcludeAFlag: skipExcludeAFlag,
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
	vm := defaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	ctx := HintRunnerContext{
		ExcludedArc: 0,
	}

	var skipExcludeBMinusA ApCellRef = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		hint := AssertLeIsSecondArcExcluded{
			skipExcludeBMinusA: skipExcludeBMinusA,
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
	vm := defaultVirtualMachine()

	rand := defaultRandGenerator()
	ctx := HintRunnerContext{
		ExcludedArc: 0,
	}

	rangeCheckPtr := vm.Memory.AllocateBuiltinSegment(&builtins.RangeCheck{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// store the range check ptr at current ap
		writeTo(
			vm,
			VM.ExecutionSegment,
			vm.Context.Ap,
			memory.MemoryValueFromMemoryAddress(&rangeCheckPtr),
		)

		r1 := randomFeltElement(rand)
		r2 := randomFeltElement(rand)
		hint := AssertLeFindSmallArc{
			a:             Immediate(r1),
			b:             Immediate(r2),
			rangeCheckPtr: Deref{ApCellRef(0)},
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
	vm := defaultVirtualMachine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		hint := RandomEcPoint{
			x: ApCellRef(0),
			y: ApCellRef(1),
		}

		err := hint.Execute(vm)
		if err != nil {
			b.Error(err)
			break
		}

		vm.Context.Ap += 2
	}

}
