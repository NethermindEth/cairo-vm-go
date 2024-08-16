package core

import (
	"io"
	"math/big"
	"os"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestAllocSegment(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 3
	vm.Context.Fp = 0

	var ap hinter.ApCellRef = 5
	var fp hinter.FpCellRef = 9

	alloc1 := AllocSegment{ap}
	alloc2 := AllocSegment{fp}

	err := alloc1.Execute(vm, nil)
	require.Nil(t, err)
	require.Equal(t, 3, len(vm.Memory.Segments))
	require.Equal(
		t,
		mem.MemoryValueFromSegmentAndOffset(2, 0),
		utils.ReadFrom(vm, VM.ExecutionSegment, vm.Context.Ap+5),
	)

	err = alloc2.Execute(vm, nil)
	require.Nil(t, err)
	require.Equal(t, 4, len(vm.Memory.Segments))
	require.Equal(
		t,
		mem.MemoryValueFromSegmentAndOffset(3, 0),
		utils.ReadFrom(vm, VM.ExecutionSegment, vm.Context.Fp+9),
	)

}

func TestTestLessThanTrue(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	utils.WriteTo(vm, VM.ExecutionSegment, 0, mem.MemoryValueFromInt(23))

	var dst hinter.ApCellRef = 1
	var rhsRef hinter.FpCellRef = 0
	rhs := hinter.Deref{Deref: rhsRef}

	lhs := hinter.Immediate(f.NewElement(13))

	hint := TestLessThan{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm, nil)
	require.NoError(t, err)
	require.Equal(
		t,
		mem.MemoryValueFromInt(1),
		utils.ReadFrom(vm, VM.ExecutionSegment, 1),
		"Expected the hint to evaluate to True when lhs is less than rhs",
	)
}

func TestTestLessThanFalse(t *testing.T) {
	testCases := []struct {
		lhsValue    f.Element
		expectedMsg string
	}{
		{f.NewElement(32), "Expected the hint to evaluate to False when lhs is larger"},
		{f.NewElement(17), "Expected the hint to evaluate to False when values are equal"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedMsg, func(t *testing.T) {
			vm := VM.DefaultVirtualMachine()
			vm.Context.Ap = 0
			vm.Context.Fp = 0
			utils.WriteTo(vm, VM.ExecutionSegment, 0, mem.MemoryValueFromInt(17))

			var dst hinter.ApCellRef = 1
			var rhsRef hinter.FpCellRef = 0
			rhs := hinter.Deref{Deref: rhsRef}

			lhs := hinter.Immediate(tc.lhsValue)
			hint := TestLessThan{
				dst: dst,
				lhs: lhs,
				rhs: rhs,
			}

			err := hint.Execute(vm, nil)
			require.NoError(t, err)
			require.Equal(
				t,
				mem.EmptyMemoryValueAsFelt(),
				utils.ReadFrom(vm, VM.ExecutionSegment, 1),
				tc.expectedMsg,
			)
		})
	}
}

func TestTestLessThanOrEqTrue(t *testing.T) {
	testCases := []struct {
		lhsValue    f.Element
		expectedMsg string
	}{
		{f.NewElement(13), "Expected the hint to evaluate to True when lhs is less than rhs"},
		{f.NewElement(23), "Expected the hint to evaluate to True when values are equal"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedMsg, func(t *testing.T) {
			vm := VM.DefaultVirtualMachine()
			vm.Context.Ap = 0
			vm.Context.Fp = 0
			utils.WriteTo(vm, VM.ExecutionSegment, 0, mem.MemoryValueFromInt(23))

			var dst hinter.ApCellRef = 1
			var rhsRef hinter.FpCellRef = 0
			rhs := hinter.Deref{Deref: rhsRef}

			lhs := hinter.Immediate(tc.lhsValue)
			hint := TestLessThanOrEqual{
				dst: dst,
				lhs: lhs,
				rhs: rhs,
			}

			err := hint.Execute(vm, nil)
			require.NoError(t, err)
			require.Equal(
				t,
				mem.MemoryValueFromInt(1),
				utils.ReadFrom(vm, VM.ExecutionSegment, 1),
				tc.expectedMsg,
			)
		})
	}
}

func TestTestLessThanOrEqFalse(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	utils.WriteTo(vm, VM.ExecutionSegment, 0, mem.MemoryValueFromInt(17))

	var dst hinter.ApCellRef = 1
	var rhsRef hinter.FpCellRef = 0
	rhs := hinter.Deref{Deref: rhsRef}

	lhs := hinter.Immediate(f.NewElement(32))

	hint := TestLessThanOrEqual{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm, nil)
	require.NoError(t, err)
	require.Equal(
		t,
		mem.EmptyMemoryValueAsFelt(),
		utils.ReadFrom(vm, VM.ExecutionSegment, 1),
		"Expected the hint to evaluate to False when lhs is larger",
	)
}

func TestLessThanOrEqualAddressTrue(t *testing.T) {
	// Address of lhs and rhs are same (SegmentIndex and Offset)
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	addr := mem.MemoryAddress{
		SegmentIndex: VM.ExecutionSegment,
		Offset:       23,
	}
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, mem.MemoryValueFromMemoryAddress(&addr))

	var dst hinter.ApCellRef = 1
	rhs := hinter.Deref{Deref: hinter.ApCellRef(0)}
	lhs := hinter.Deref{Deref: hinter.ApCellRef(0)}

	hint := TestLessThanOrEqualAddress{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm, nil)
	require.NoError(t, err)
	require.Equal(
		t,
		mem.MemoryValueFromInt(1),
		utils.ReadFrom(vm, VM.ExecutionSegment, 1),
	)

	// Address of lhs is less than the address of rhs (Offset)
	vm = VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	rhsAddr := mem.MemoryAddress{
		SegmentIndex: VM.ExecutionSegment,
		Offset:       23,
	}
	lhsAddr := mem.MemoryAddress{
		SegmentIndex: VM.ExecutionSegment,
		Offset:       17,
	}
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, mem.MemoryValueFromMemoryAddress(&rhsAddr))
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap+1, mem.MemoryValueFromMemoryAddress(&lhsAddr))

	dst = 2
	rhs = hinter.Deref{Deref: hinter.ApCellRef(0)}
	lhs = hinter.Deref{Deref: hinter.ApCellRef(1)}

	hint = TestLessThanOrEqualAddress{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err = hint.Execute(vm, nil)
	require.NoError(t, err)
	require.Equal(
		t,
		mem.MemoryValueFromInt(1),
		utils.ReadFrom(vm, VM.ExecutionSegment, 2),
	)

	// Address of lhs is less than the address of rhs (SegmentIndex)
	vm = VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	rhsAddr = mem.MemoryAddress{
		SegmentIndex: VM.ExecutionSegment + 1,
		Offset:       17,
	}
	lhsAddr = mem.MemoryAddress{
		SegmentIndex: VM.ExecutionSegment,
		Offset:       23,
	}
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, mem.MemoryValueFromMemoryAddress(&rhsAddr))
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap+1, mem.MemoryValueFromMemoryAddress(&lhsAddr))

	dst = 2
	rhs = hinter.Deref{Deref: hinter.ApCellRef(0)}
	lhs = hinter.Deref{Deref: hinter.ApCellRef(1)}

	hint = TestLessThanOrEqualAddress{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err = hint.Execute(vm, nil)
	require.NoError(t, err)
	require.Equal(
		t,
		mem.MemoryValueFromInt(1),
		utils.ReadFrom(vm, VM.ExecutionSegment, 2),
	)
}

func TestLessThanOrEqualAddressFalse(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	rhsAddr := mem.MemoryAddress{
		SegmentIndex: VM.ExecutionSegment,
		Offset:       17,
	}
	lhsAddr := mem.MemoryAddress{
		SegmentIndex: VM.ExecutionSegment,
		Offset:       23,
	}
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, mem.MemoryValueFromMemoryAddress(&rhsAddr))
	utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap+1, mem.MemoryValueFromMemoryAddress(&lhsAddr))

	var dst hinter.ApCellRef = 2
	rhs := hinter.Deref{Deref: hinter.ApCellRef(0)}
	lhs := hinter.Deref{Deref: hinter.ApCellRef(1)}

	hint := TestLessThanOrEqualAddress{
		dst: dst,
		lhs: lhs,
		rhs: rhs,
	}

	err := hint.Execute(vm, nil)
	require.NoError(t, err)
	require.Equal(
		t,
		mem.EmptyMemoryValueAsFelt(),
		utils.ReadFrom(vm, VM.ExecutionSegment, 2),
		"Expected the hint to evaluate to False when address of lhs is larger",
	)
}

func TestLinearSplit(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	value := hinter.Immediate(f.NewElement(42*223344 + 14))
	scalar := hinter.Immediate(f.NewElement(42))
	maxX := hinter.Immediate(f.NewElement(9999999999))
	var x hinter.ApCellRef = 0
	var y hinter.ApCellRef = 1

	hint := LinearSplit{
		value:  value,
		scalar: scalar,
		maxX:   maxX,
		x:      x,
		y:      y,
	}

	err := hint.Execute(vm, nil)
	require.NoError(t, err)
	xx := utils.ReadFrom(vm, VM.ExecutionSegment, 0)
	require.Equal(t, xx, mem.MemoryValueFromInt(223344))
	yy := utils.ReadFrom(vm, VM.ExecutionSegment, 1)
	require.Equal(t, yy, mem.MemoryValueFromInt(14))

	vm = VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	//Lower max_x
	maxX = hinter.Immediate(f.NewElement(223343))
	hint = LinearSplit{
		value:  value,
		scalar: scalar,
		maxX:   maxX,
		x:      x,
		y:      y,
	}

	err = hint.Execute(vm, nil)
	require.NoError(t, err)
	xx = utils.ReadFrom(vm, VM.ExecutionSegment, 0)
	require.Equal(t, xx, mem.MemoryValueFromInt(223343))
	yy = utils.ReadFrom(vm, VM.ExecutionSegment, 1)
	require.Equal(t, yy, mem.MemoryValueFromInt(14+42))
}

func TestWideMul128(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstLow hinter.ApCellRef = 1
	var dstHigh hinter.ApCellRef = 2

	lhsBytes := new(uint256.Int).Lsh(uint256.NewInt(1), 127).Bytes32()
	lhsFelt, err := f.BigEndian.Element(&lhsBytes)
	require.NoError(t, err)

	rhsFelt := f.NewElement(1<<8 + 1)

	lhs := hinter.Immediate(lhsFelt)
	rhs := hinter.Immediate(rhsFelt)

	hint := WideMul128{
		low:  dstLow,
		high: dstHigh,
		lhs:  lhs,
		rhs:  rhs,
	}

	err = hint.Execute(vm, nil)
	require.Nil(t, err)

	low := &f.Element{}
	low.SetBigInt(big.NewInt(1).Lsh(big.NewInt(1), 127))

	require.Equal(
		t,
		mem.MemoryValueFromFieldElement(low),
		utils.ReadFrom(vm, VM.ExecutionSegment, 1),
	)
	require.Equal(
		t,
		mem.MemoryValueFromInt(1<<7),
		utils.ReadFrom(vm, VM.ExecutionSegment, 2),
	)
}

func TestDivMod(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var quo hinter.ApCellRef = 1
	var rem hinter.ApCellRef = 2

	lhsValue := hinter.Immediate(f.NewElement(89))
	rhsValue := hinter.Immediate(f.NewElement(7))

	hint := DivMod{
		lhs:       lhsValue,
		rhs:       rhsValue,
		quotient:  quo,
		remainder: rem,
	}

	err := hint.Execute(vm, nil)
	require.Nil(t, err)

	expectedQuotient := mem.MemoryValueFromInt(12)
	expectedRemainder := mem.MemoryValueFromInt(5)

	actualQuotient := utils.ReadFrom(vm, VM.ExecutionSegment, 1)
	actualRemainder := utils.ReadFrom(vm, VM.ExecutionSegment, 2)

	require.Equal(t, expectedQuotient, actualQuotient)
	require.Equal(t, expectedRemainder, actualRemainder)
}

func TestDivModDivisionByZeroError(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var quo hinter.ApCellRef = 1
	var rem hinter.ApCellRef = 2

	lhsValue := hinter.Immediate(f.NewElement(43))
	rhsValue := hinter.Immediate(f.NewElement(0))

	hint := DivMod{
		lhs:       lhsValue,
		rhs:       rhsValue,
		quotient:  quo,
		remainder: rem,
	}

	err := hint.Execute(vm, nil)
	require.ErrorContains(t, err, "cannot be divided by zero, rhs: 0")
}

func TestUint256DivMod(t *testing.T) {
	t.Run("test uint256DivMod", func(t *testing.T) {
		vm := VM.DefaultVirtualMachine()
		vm.Context.Ap = 0
		vm.Context.Fp = 0

		var quotient0 hinter.ApCellRef = 1
		var quotient1 hinter.ApCellRef = 2
		var remainder0 hinter.ApCellRef = 3
		var remainder1 hinter.ApCellRef = 4

		dividend0Felt := f.NewElement(89)
		dividend1Felt := f.NewElement(72)

		divisor0Felt := f.NewElement(3)
		divisor1Felt := f.NewElement(7)

		hint := Uint256DivMod{
			dividend0:  hinter.Immediate(dividend0Felt),
			dividend1:  hinter.Immediate(dividend1Felt),
			divisor0:   hinter.Immediate(divisor0Felt),
			divisor1:   hinter.Immediate(divisor1Felt),
			quotient0:  quotient0,
			quotient1:  quotient1,
			remainder0: remainder0,
			remainder1: remainder1,
		}

		err := hint.Execute(vm, nil)
		require.Nil(t, err)

		quotient0Val := &f.Element{}
		_, err = quotient0Val.SetString("10")
		require.Nil(t, err)

		require.Equal(
			t,
			mem.MemoryValueFromFieldElement(quotient0Val),
			utils.ReadFrom(vm, VM.ExecutionSegment, 1),
		)

		quotient1Val := &f.Element{}
		quotient1Val.SetZero()
		require.Nil(t, err)

		require.Equal(
			t,
			mem.MemoryValueFromFieldElement(quotient1Val),
			utils.ReadFrom(vm, VM.ExecutionSegment, 2),
		)

		remainder0Val := &f.Element{}
		_, err = remainder0Val.SetString("59")
		require.Nil(t, err)

		require.Equal(
			t,
			mem.MemoryValueFromFieldElement(remainder0Val),
			utils.ReadFrom(vm, VM.ExecutionSegment, 3),
		)

		remainder1Val := &f.Element{}
		_, err = remainder1Val.SetString("2")
		require.Nil(t, err)

		require.Equal(
			t,
			mem.MemoryValueFromFieldElement(remainder1Val),
			utils.ReadFrom(vm, VM.ExecutionSegment, 4),
		)
	})
	t.Run("test uint256DivMod with 256-bit numbers;", func(t *testing.T) {
		vm := VM.DefaultVirtualMachine()
		vm.Context.Ap = 0
		vm.Context.Fp = 0

		var quotient0 hinter.ApCellRef = 1
		var quotient1 hinter.ApCellRef = 2
		var remainder0 hinter.ApCellRef = 3
		var remainder1 hinter.ApCellRef = 4

		b := new(uint256.Int).Lsh(uint256.NewInt(1), 127).Bytes32()

		dividend0Felt, err := f.BigEndian.Element(&b)
		require.NoError(t, err)
		dividend1Felt := f.NewElement(1<<8 + 1)

		divisor0Felt := f.NewElement(1<<8 + 1)
		divisor1Felt := f.NewElement(1<<8 + 1)

		hint := Uint256DivMod{
			dividend0:  hinter.Immediate(dividend0Felt),
			dividend1:  hinter.Immediate(dividend1Felt),
			divisor0:   hinter.Immediate(divisor0Felt),
			divisor1:   hinter.Immediate(divisor1Felt),
			quotient0:  quotient0,
			quotient1:  quotient1,
			remainder0: remainder0,
			remainder1: remainder1,
		}

		err = hint.Execute(vm, nil)
		require.Nil(t, err)

		quotient0Val := &f.Element{}
		quotient0Val.SetOne()
		require.Nil(t, err)

		require.Equal(
			t,
			mem.MemoryValueFromFieldElement(quotient0Val),
			utils.ReadFrom(vm, VM.ExecutionSegment, 1),
		)

		quotient1Val := &f.Element{}
		quotient1Val.SetZero()
		require.Nil(t, err)

		require.Equal(
			t,
			mem.MemoryValueFromFieldElement(quotient1Val),
			utils.ReadFrom(vm, VM.ExecutionSegment, 2),
		)

		remainder0Val := &f.Element{}
		_, err = remainder0Val.SetString("170141183460469231731687303715884105471")
		require.Nil(t, err)

		require.Equal(
			t,
			mem.MemoryValueFromFieldElement(remainder0Val),
			utils.ReadFrom(vm, VM.ExecutionSegment, 3),
		)

		remainder1Val := &f.Element{}
		remainder1Val.SetZero()
		require.Nil(t, err)

		require.Equal(
			t,
			mem.MemoryValueFromFieldElement(remainder1Val),
			utils.ReadFrom(vm, VM.ExecutionSegment, 4),
		)
	})
}

func TestUint256DivModDivisionByZero(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstQuotient0 hinter.ApCellRef = 1
	var dstQuotient1 hinter.ApCellRef = 2
	var dstRemainder0 hinter.ApCellRef = 3
	var dstRemainder1 hinter.ApCellRef = 4

	dividend0Felt := f.NewElement(1<<8 + 1)
	dividend1Felt := f.NewElement(1<<8 + 1)

	divisor0Felt := f.NewElement(0)
	divisor1Felt := f.NewElement(0)

	hint := Uint256DivMod{
		dividend0:  hinter.Immediate(dividend0Felt),
		dividend1:  hinter.Immediate(dividend1Felt),
		divisor0:   hinter.Immediate(divisor0Felt),
		divisor1:   hinter.Immediate(divisor1Felt),
		quotient0:  dstQuotient0,
		quotient1:  dstQuotient1,
		remainder0: dstRemainder0,
		remainder1: dstRemainder1,
	}

	err := hint.Execute(vm, nil)
	require.ErrorContains(t, err, "cannot be divided by zero, divisor: 0")
}

func TestWideMul128IncorrectRange(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstLow hinter.ApCellRef = 1
	var dstHigh hinter.ApCellRef = 2

	lhsBytes := new(uint256.Int).Lsh(uint256.NewInt(1), 128).Bytes32()
	lhsFelt, err := f.BigEndian.Element(&lhsBytes)
	require.NoError(t, err)

	lhs := hinter.Immediate(lhsFelt)
	rhs := hinter.Immediate(f.NewElement(1))

	hint := WideMul128{
		low:  dstLow,
		high: dstHigh,
		lhs:  lhs,
		rhs:  rhs,
	}

	err = hint.Execute(vm, nil)
	require.ErrorContains(t, err, "should be u128")
}

func TestDebugPrint(t *testing.T) {
	//Save the old stdout
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	utils.WriteTo(vm, VM.ExecutionSegment, 0, mem.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, 2))
	utils.WriteTo(vm, VM.ExecutionSegment, 1, mem.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, 5))
	utils.WriteTo(vm, VM.ExecutionSegment, 2, mem.MemoryValueFromInt(10))
	utils.WriteTo(vm, VM.ExecutionSegment, 3, mem.MemoryValueFromInt(20))
	utils.WriteTo(vm, VM.ExecutionSegment, 4, mem.MemoryValueFromInt(30))

	var starRef hinter.ApCellRef = 0
	var endRef hinter.ApCellRef = 1
	start := hinter.Deref{Deref: starRef}
	end := hinter.Deref{Deref: endRef}
	hint := DebugPrint{
		start: start,
		end:   end,
	}
	expected := []byte("[DEBUG] a\n[DEBUG] 14\n[DEBUG] 1e\n")
	err := hint.Execute(vm, nil)

	w.Close()
	out, _ := io.ReadAll(r)
	//Restore stdout at the end of the test
	os.Stdout = rescueStdout

	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestSquareRoot(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0
	var dst hinter.ApCellRef = 1

	value := hinter.Immediate(f.NewElement(36))
	hint := SquareRoot{
		value: value,
		dst:   dst,
	}

	err := hint.Execute(vm, nil)

	require.NoError(t, err)
	require.Equal(
		t,
		mem.MemoryValueFromInt(6),
		utils.ReadFrom(vm, VM.ExecutionSegment, 1),
	)

	dst = 2
	value = hinter.Immediate(f.NewElement(30))
	hint = SquareRoot{
		value: value,
		dst:   dst,
	}

	err = hint.Execute(vm, nil)

	require.NoError(t, err)
	require.Equal(
		t,
		mem.MemoryValueFromInt(5),
		utils.ReadFrom(vm, VM.ExecutionSegment, 2),
	)
}

func TestUint256SquareRootLow(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var sqrt0 hinter.ApCellRef = 1
	var sqrt1 hinter.ApCellRef = 2
	var remainderLow hinter.ApCellRef = 3
	var remainderHigh hinter.ApCellRef = 4
	var sqrtMul2MinusRemainderGeU128 hinter.ApCellRef = 5

	valueLow := hinter.Immediate(f.NewElement(121))
	valueHigh := hinter.Immediate(f.NewElement(0))

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

	require.NoError(t, err)

	expectedSqrt0 := mem.MemoryValueFromInt(11)
	expectedSqrt1 := mem.MemoryValueFromInt(0)
	expectedRemainderLow := mem.MemoryValueFromInt(0)
	expectedRemainderHigh := mem.MemoryValueFromInt(0)
	expectedSqrtMul2MinusRemainderGeU128 := mem.MemoryValueFromInt(0)

	actualSqrt0 := utils.ReadFrom(vm, VM.ExecutionSegment, 1)
	actualSqrt1 := utils.ReadFrom(vm, VM.ExecutionSegment, 2)
	actualRemainderLow := utils.ReadFrom(vm, VM.ExecutionSegment, 3)
	actualRemainderHigh := utils.ReadFrom(vm, VM.ExecutionSegment, 4)
	actualSqrtMul2MinusRemainderGeU128 := utils.ReadFrom(vm, VM.ExecutionSegment, 5)

	require.Equal(t, expectedSqrt0, actualSqrt0)
	require.Equal(t, expectedSqrt1, actualSqrt1)
	require.Equal(t, expectedRemainderLow, actualRemainderLow)
	require.Equal(t, expectedRemainderHigh, actualRemainderHigh)
	require.Equal(t, expectedSqrtMul2MinusRemainderGeU128, actualSqrtMul2MinusRemainderGeU128)
}

func TestUint256SquareRootHigh(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var sqrt0 hinter.ApCellRef = 1
	var sqrt1 hinter.ApCellRef = 2
	var remainderLow hinter.ApCellRef = 3
	var remainderHigh hinter.ApCellRef = 4
	var sqrtMul2MinusRemainderGeU128 hinter.ApCellRef = 5

	valueLow := hinter.Immediate(f.NewElement(0))
	valueHigh := hinter.Immediate(f.NewElement(1 << 8))

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

	require.NoError(t, err)

	expectedSqrt0 := mem.MemoryValueFromInt(0)
	expectedSqrt1 := mem.MemoryValueFromInt(16)
	expectedRemainderLow := mem.MemoryValueFromInt(0)
	expectedRemainderHigh := mem.MemoryValueFromInt(0)
	expectedSqrtMul2MinusRemainderGeU128 := mem.MemoryValueFromInt(0)

	actualSqrt0 := utils.ReadFrom(vm, VM.ExecutionSegment, 1)
	actualSqrt1 := utils.ReadFrom(vm, VM.ExecutionSegment, 2)
	actualRemainderLow := utils.ReadFrom(vm, VM.ExecutionSegment, 3)
	actualRemainderHigh := utils.ReadFrom(vm, VM.ExecutionSegment, 4)
	actualSqrtMul2MinusRemainderGeU128 := utils.ReadFrom(vm, VM.ExecutionSegment, 5)

	require.Equal(t, expectedSqrt0, actualSqrt0)
	require.Equal(t, expectedSqrt1, actualSqrt1)
	require.Equal(t, expectedRemainderLow, actualRemainderLow)
	require.Equal(t, expectedRemainderHigh, actualRemainderHigh)
	require.Equal(t, expectedSqrtMul2MinusRemainderGeU128, actualSqrtMul2MinusRemainderGeU128)
}

func TestUint256SquareRoot(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var sqrt0 hinter.ApCellRef = 1
	var sqrt1 hinter.ApCellRef = 2
	var remainderLow hinter.ApCellRef = 3
	var remainderHigh hinter.ApCellRef = 4
	var sqrtMul2MinusRemainderGeU128 hinter.ApCellRef = 5

	valueLow := hinter.Immediate(f.NewElement(51))
	valueHigh := hinter.Immediate(f.NewElement(1024))

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

	require.NoError(t, err)

	expectedSqrt0 := mem.MemoryValueFromInt(0)
	expectedSqrt1 := mem.MemoryValueFromInt(32)
	expectedRemainderLow := mem.MemoryValueFromInt(51)
	expectedRemainderHigh := mem.MemoryValueFromInt(0)
	expectedSqrtMul2MinusRemainderGeU128 := mem.MemoryValueFromInt(0)

	actualSqrt0 := utils.ReadFrom(vm, VM.ExecutionSegment, 1)
	actualSqrt1 := utils.ReadFrom(vm, VM.ExecutionSegment, 2)
	actualRemainderLow := utils.ReadFrom(vm, VM.ExecutionSegment, 3)
	actualRemainderHigh := utils.ReadFrom(vm, VM.ExecutionSegment, 4)
	actualSqrtMul2MinusRemainderGeU128 := utils.ReadFrom(vm, VM.ExecutionSegment, 5)

	require.Equal(t, expectedSqrt0, actualSqrt0)
	require.Equal(t, expectedSqrt1, actualSqrt1)
	require.Equal(t, expectedRemainderLow, actualRemainderLow)
	require.Equal(t, expectedRemainderHigh, actualRemainderHigh)
	require.Equal(t, expectedSqrtMul2MinusRemainderGeU128, actualSqrtMul2MinusRemainderGeU128)
}

func TestUint512DivModByUint256(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstQuotient0 hinter.ApCellRef = 1
	var dstQuotient1 hinter.ApCellRef = 2
	var dstQuotient2 hinter.ApCellRef = 3
	var dstQuotient3 hinter.ApCellRef = 4
	var dstRemainder0 hinter.ApCellRef = 5
	var dstRemainder1 hinter.ApCellRef = 6

	b := new(uint256.Int).Lsh(uint256.NewInt(1), 127).Bytes32()

	dividend0Felt, err := f.BigEndian.Element(&b)
	require.NoError(t, err)
	dividend1Felt := f.NewElement(1<<8 + 1)
	dividend2Felt, err := f.BigEndian.Element(&b)
	require.NoError(t, err)
	dividend3Felt := f.NewElement(1<<8 + 1)

	divisor0Felt := f.NewElement(1<<8 + 1)
	divisor1Felt := f.NewElement(1<<8 + 1)

	hint := Uint512DivModByUint256{
		dividend0:  hinter.Immediate(dividend0Felt),
		dividend1:  hinter.Immediate(dividend1Felt),
		dividend2:  hinter.Immediate(dividend2Felt),
		dividend3:  hinter.Immediate(dividend3Felt),
		divisor0:   hinter.Immediate(divisor0Felt),
		divisor1:   hinter.Immediate(divisor1Felt),
		quotient0:  dstQuotient0,
		quotient1:  dstQuotient1,
		quotient2:  dstQuotient2,
		quotient3:  dstQuotient3,
		remainder0: dstRemainder0,
		remainder1: dstRemainder1,
	}

	err = hint.Execute(vm, nil)
	require.Nil(t, err)

	quotient0 := &f.Element{}
	_, err = quotient0.SetString("170141183460469231731687303715884105730")
	require.Nil(t, err)

	require.Equal(
		t,
		mem.MemoryValueFromFieldElement(quotient0),
		utils.ReadFrom(vm, VM.ExecutionSegment, 1),
	)

	quotient1 := &f.Element{}
	_, err = quotient1.SetString("662027951208051485337304683719393406")
	require.Nil(t, err)

	require.Equal(
		t,
		mem.MemoryValueFromFieldElement(quotient1),
		utils.ReadFrom(vm, VM.ExecutionSegment, 2),
	)

	quotient2 := &f.Element{}
	quotient2.SetOne()

	require.Equal(
		t,
		mem.MemoryValueFromFieldElement(quotient2),
		utils.ReadFrom(vm, VM.ExecutionSegment, 3),
	)

	quotient3 := &f.Element{}
	quotient3.SetZero()

	require.Equal(
		t,
		mem.MemoryValueFromFieldElement(quotient3),
		utils.ReadFrom(vm, VM.ExecutionSegment, 4),
	)

	remainder0 := &f.Element{}
	_, err = remainder0.SetString("340282366920938463463374607431768210942")
	require.Nil(t, err)

	require.Equal(
		t,
		mem.MemoryValueFromFieldElement(remainder0),
		utils.ReadFrom(vm, VM.ExecutionSegment, 5),
	)

	remainder1 := &f.Element{}
	remainder1.SetZero()

	require.Equal(
		t,
		mem.MemoryValueFromFieldElement(remainder1),
		utils.ReadFrom(vm, VM.ExecutionSegment, 6),
	)
}

func TestUint512DivModByUint256DivisionByZero(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	var dstQuotient0 hinter.ApCellRef = 1
	var dstQuotient1 hinter.ApCellRef = 2
	var dstQuotient2 hinter.ApCellRef = 3
	var dstQuotient3 hinter.ApCellRef = 4
	var dstRemainder0 hinter.ApCellRef = 5
	var dstRemainder1 hinter.ApCellRef = 6

	b := new(uint256.Int).Lsh(uint256.NewInt(1), 127).Bytes32()

	dividend0Felt, err := f.BigEndian.Element(&b)
	require.NoError(t, err)
	dividend1Felt := f.NewElement(1<<8 + 1)
	dividend2Felt, err := f.BigEndian.Element(&b)
	require.NoError(t, err)
	dividend3Felt := f.NewElement(1<<8 + 1)

	divisor0Felt := f.NewElement(0)
	divisor1Felt := f.NewElement(0)

	hint := Uint512DivModByUint256{
		dividend0:  hinter.Immediate(dividend0Felt),
		dividend1:  hinter.Immediate(dividend1Felt),
		dividend2:  hinter.Immediate(dividend2Felt),
		dividend3:  hinter.Immediate(dividend3Felt),
		divisor0:   hinter.Immediate(divisor0Felt),
		divisor1:   hinter.Immediate(divisor1Felt),
		quotient0:  dstQuotient0,
		quotient1:  dstQuotient1,
		quotient2:  dstQuotient2,
		quotient3:  dstQuotient3,
		remainder0: dstRemainder0,
		remainder1: dstRemainder1,
	}

	err = hint.Execute(vm, nil)
	require.ErrorContains(t, err, "division by zero")
}

func TestAllocConstantSize(t *testing.T) {
	vm := VM.DefaultVirtualMachine()

	sizes := [3]hinter.Immediate{
		hinter.Immediate(f.NewElement(15)),
		hinter.Immediate(f.NewElement(13)),
		hinter.Immediate(f.NewElement(2)),
	}
	expectedAddrs := [3]mem.MemoryAddress{
		{SegmentIndex: 2, Offset: 0},
		{SegmentIndex: 2, Offset: 15},
		{SegmentIndex: 2, Offset: 28},
	}

	ctx := hinter.HintRunnerContext{
		ConstantSizeSegment: mem.UnknownAddress,
	}

	for i := 0; i < len(sizes); i++ {
		hint := AllocConstantSize{
			Dst:  hinter.ApCellRef(i),
			Size: sizes[i],
		}

		err := hint.Execute(vm, &ctx)
		require.NoError(t, err)

		val := utils.ReadFrom(vm, 1, uint64(i))
		ptr, err := val.MemoryAddress()
		require.NoError(t, err)

		require.Equal(t, &expectedAddrs[i], ptr)
	}

	require.Equal(t, ctx.ConstantSizeSegment, mem.MemoryAddress{SegmentIndex: 2, Offset: 30})
}

func TestAssertLeFindSmallArc(t *testing.T) {
	testCases := []struct {
		aFelt, bFelt                    f.Element
		expectedRem1, expectedQuotient1 mem.MemoryValue
		expectedRem2, expectedQuotient2 mem.MemoryValue
		expectedExcludedArc             int
	}{
		// First test case
		{
			aFelt:               f.NewElement(1024),
			bFelt:               f.NewElement(1025),
			expectedRem1:        mem.MemoryValueFromInt(1),
			expectedQuotient1:   mem.MemoryValueFromInt(0),
			expectedRem2:        mem.MemoryValueFromInt(1024),
			expectedQuotient2:   mem.MemoryValueFromInt(0),
			expectedExcludedArc: 2,
		},
		// Second test case
		{
			// 2974197561122951277584414786853691079
			aFelt: f.Element{
				13984218141608664100,
				13287333742236603547,
				18446744073709551615,
				229878458336812643,
			},
			// 306150973282131698343156044521811432643
			bFelt: f.Element{
				6079377935050068685,
				3868297591914914705,
				18446744073709551587,
				162950233538363292,
			},
			// 2974197561122951277584414786853691079
			expectedRem1: mem.MemoryValueFromFieldElement(
				&f.Element{
					13984218141608664100,
					13287333742236603547,
					18446744073709551615,
					229878458336812643,
				}),
			expectedQuotient1: mem.MemoryValueFromInt(0),
			// 112792682047919106056116278761420227
			expectedRem2: mem.MemoryValueFromFieldElement(
				&f.Element{
					10541903867150958026,
					18251079960242638581,
					18446744073709551615,
					509532527505005161,
				}),
			expectedQuotient2:   mem.MemoryValueFromInt(57),
			expectedExcludedArc: 2,
		},
	}

	for _, tc := range testCases {
		// Need to create a new VM for each test case
		// to avoid rewriting in same memory address error
		vm := VM.DefaultVirtualMachine()
		vm.Context.Ap = 0
		vm.Context.Fp = 0
		// The addr that the range check pointer will point to
		addr := vm.Memory.AllocateBuiltinSegment(&builtins.RangeCheck{RANGE_CHECK_N_PARTS: 8})
		utils.WriteTo(vm, VM.ExecutionSegment, vm.Context.Ap, mem.MemoryValueFromMemoryAddress(&addr))

		hint := AssertLeFindSmallArc{
			A:             hinter.Immediate(tc.aFelt),
			B:             hinter.Immediate(tc.bFelt),
			RangeCheckPtr: hinter.Deref{Deref: hinter.ApCellRef(0)},
		}

		ctx := hinter.SetContextWithScope(map[string]any{"excluded": 0})

		err := hint.Execute(vm, ctx)

		require.NoError(t, err)

		expectedPtr := mem.MemoryValueFromMemoryAddress(&addr)

		actualRem1 := utils.ReadFrom(vm, 2, 0)
		actualQuotient1 := utils.ReadFrom(vm, 2, 1)
		actualRem2 := utils.ReadFrom(vm, 2, 2)
		actualQuotient2 := utils.ReadFrom(vm, 2, 3)
		actual1Ptr := utils.ReadFrom(vm, 1, 0)
		actualExcludedArc, err := ctx.ScopeManager.GetVariableValue("excluded")

		require.NoError(t, err)

		require.Equal(t, tc.expectedRem1, actualRem1)
		require.Equal(t, tc.expectedQuotient1, actualQuotient1)
		require.Equal(t, tc.expectedRem2, actualRem2)
		require.Equal(t, tc.expectedQuotient2, actualQuotient2)
		require.Equal(t, expectedPtr, actual1Ptr)
		require.Equal(t, tc.expectedExcludedArc, actualExcludedArc)
	}
}

func TestAssertLeIsFirstArcExcluded(t *testing.T) {
	vm := VM.DefaultVirtualMachine()

	ctx := hinter.SetContextWithScope(map[string]any{"excluded": 2})
	var flag hinter.ApCellRef = 0

	hint := AssertLeIsFirstArcExcluded{
		SkipExcludeAFlag: flag,
	}

	err := hint.Execute(vm, ctx)
	require.NoError(t, err)

	expected := mem.MemoryValueFromInt(1)

	actual := utils.ReadFrom(vm, VM.ExecutionSegment, 0)

	require.Equal(t, expected, actual)
}

func TestAssertLeIsSecondArcExcluded(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	ctx := hinter.SetContextWithScope(map[string]any{"excluded": 1})
	var flag hinter.ApCellRef = 0

	hint := AssertLeIsSecondArcExcluded{
		SkipExcludeBMinusA: flag,
	}

	err := hint.Execute(vm, ctx)
	require.NoError(t, err)

	expected := mem.MemoryValueFromInt(0)
	actual := utils.ReadFrom(vm, VM.ExecutionSegment, 0)
	require.Equal(t, expected, actual)
}

func TestRandomEcPoint(t *testing.T) {
	vm := VM.DefaultVirtualMachine()
	vm.Context.Ap = 0
	vm.Context.Fp = 0

	hint := RandomEcPoint{
		x: hinter.ApCellRef(0),
		y: hinter.ApCellRef(1),
	}

	err := hint.Execute(vm)
	require.NoError(t, err)

	expectedX := mem.MemoryValueFromFieldElement(
		&f.Element{12217889558999792019, 3067322962467879919, 3160430244162662030, 474947714424245026},
	)
	expectedY := mem.MemoryValueFromFieldElement(
		&f.Element{12193331470568888984, 1737428559173019240, 11500517745011090163, 245183001587853482},
	)

	actualX := utils.ReadFrom(vm, VM.ExecutionSegment, 0)
	actualY := utils.ReadFrom(vm, VM.ExecutionSegment, 1)

	require.Equal(t, expectedX, actualX)
	require.Equal(t, expectedY, actualY)
}

func TestFieldSqrt(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint64
		expected int
	}{
		{
			name:     "TestFieldSqrt",
			value:    49,
			expected: 7,
		},
		{
			name:     "TestFieldSqrtNonResidue",
			value:    27,
			expected: -9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vm := VM.DefaultVirtualMachine()
			vm.Context.Ap = 0
			vm.Context.Fp = 0

			value := hinter.Immediate(f.NewElement(tc.value))
			hint := FieldSqrt{
				val:  value,
				sqrt: hinter.ApCellRef(0),
			}

			err := hint.Execute(vm)

			require.NoError(t, err)
			require.Equal(
				t,
				mem.MemoryValueFromInt(tc.expected),
				utils.ReadFrom(vm, VM.ExecutionSegment, 0),
			)
		})
	}
}
