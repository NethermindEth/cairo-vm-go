package core

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/holiman/uint256"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	u "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type AllocSegment struct {
	Dst hinter.CellRefer
}

func (hint *AllocSegment) String() string {
	return "AllocSegment"
}

func (hint *AllocSegment) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	newSegment := vm.Memory.AllocateEmptySegment()
	memAddress := mem.MemoryValueFromMemoryAddress(&newSegment)

	regAddr, err := hint.Dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get register %s: %w", hint.Dst, err)
	}

	err = vm.Memory.WriteToAddress(&regAddr, &memAddress)
	if err != nil {
		return fmt.Errorf("write to address %s: %w", regAddr, err)
	}

	return nil
}

type TestLessThan struct {
	dst hinter.CellRefer
	lhs hinter.ResOperander
	rhs hinter.ResOperander
}

func (hint *TestLessThan) String() string {
	return "TestLessThan"
}

func (hint *TestLessThan) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	lhsVal, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %w", hint.lhs, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %w", hint.rhs, err)
	}

	lhsFelt, err := lhsVal.FieldElement()
	if err != nil {
		return err
	}

	rhsFelt, err := rhsVal.FieldElement()
	if err != nil {
		return err
	}

	resFelt := f.Element{}
	if lhsFelt.Cmp(rhsFelt) < 0 {
		resFelt.SetOne()
	}

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get dst address %s: %w", dstAddr, err)
	}

	mv := mem.MemoryValueFromFieldElement(&resFelt)
	err = vm.Memory.WriteToAddress(&dstAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to dst address %s: %w", dstAddr, err)
	}

	return nil
}

type TestLessThanOrEqual struct {
	dst hinter.CellRefer
	lhs hinter.ResOperander
	rhs hinter.ResOperander
}

func (hint *TestLessThanOrEqual) String() string {
	return "TestLessThanOrEqual"
}

func (hint *TestLessThanOrEqual) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	lhsVal, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %w", hint.lhs, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %w", hint.rhs, err)
	}

	lhsFelt, err := lhsVal.FieldElement()
	if err != nil {
		return err
	}

	rhsFelt, err := rhsVal.FieldElement()
	if err != nil {
		return err
	}

	resFelt := f.Element{}
	if lhsFelt.Cmp(rhsFelt) <= 0 {
		resFelt.SetOne()
	}

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get dst address %s: %w", dstAddr, err)
	}

	mv := mem.MemoryValueFromFieldElement(&resFelt)
	err = vm.Memory.WriteToAddress(&dstAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to dst address %s: %w", dstAddr, err)
	}

	return nil
}

type TestLessThanOrEqualAddress struct {
	dst hinter.CellRefer
	lhs hinter.ResOperander
	rhs hinter.ResOperander
}

func (hint *TestLessThanOrEqualAddress) String() string {
	return "TestLessThanOrEqualAddress"
}

func (hint *TestLessThanOrEqualAddress) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	lhsPtr, err := hinter.ResolveAsAddress(vm, hint.lhs)
	if err != nil {
		return fmt.Errorf("resolve lhs pointer: %w", err)
	}

	rhsPtr, err := hinter.ResolveAsAddress(vm, hint.rhs)
	if err != nil {
		return fmt.Errorf("resolve rhs pointer: %w", err)
	}

	resFelt := f.Element{}
	if lhsPtr.Cmp(rhsPtr) <= 0 {
		resFelt.SetOne()
	}

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get dst address %s: %w", dstAddr, err)
	}

	mv := mem.MemoryValueFromFieldElement(&resFelt)
	err = vm.Memory.WriteToAddress(&dstAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to dst address %s: %w", dstAddr, err)
	}

	return nil
}

type LinearSplit struct {
	value  hinter.ResOperander
	scalar hinter.ResOperander
	maxX   hinter.ResOperander
	x      hinter.CellRefer
	y      hinter.CellRefer
}

func (hint LinearSplit) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	value, err := hint.value.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve value operand %s: %w", hint.value, err)
	}
	valueField, err := value.FieldElement()
	if err != nil {
		return err
	}
	scalar, err := hint.scalar.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve scalar operand %s: %w", hint.scalar, err)
	}
	scalarField, err := scalar.FieldElement()
	if err != nil {
		return err
	}

	maxX, err := hint.maxX.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve max_x operand %s: %w", hint.maxX, err)
	}
	maxXField, err := maxX.FieldElement()
	if err != nil {
		return err
	}

	scalarBytes := scalarField.Bytes()
	valueBytes := valueField.Bytes()
	maxXBytes := maxXField.Bytes()
	scalarUint := new(uint256.Int).SetBytes(scalarBytes[:])
	valueUint := new(uint256.Int).SetBytes(valueBytes[:])
	maxXUint := new(uint256.Int).SetBytes(maxXBytes[:])

	x := (&uint256.Int{}).Div(valueUint, scalarUint)

	if x.Cmp(maxXUint) > 0 {
		x.Set(maxXUint)
	}

	y := &uint256.Int{}
	y = y.Sub(valueUint, y.Mul(scalarUint, x))

	xAddr, err := hint.x.Get(vm)
	if err != nil {
		return fmt.Errorf("get x address %s: %w", xAddr, err)
	}

	yAddr, err := hint.y.Get(vm)
	if err != nil {
		return fmt.Errorf("get y address %s: %w", yAddr, err)
	}

	xFiled := &f.Element{}
	yFiled := &f.Element{}
	xFiled.SetBytes(x.Bytes())
	yFiled.SetBytes(y.Bytes())
	mv := mem.MemoryValueFromFieldElement(xFiled)
	err = vm.Memory.WriteToAddress(&xAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to x address %s: %w", xAddr, err)
	}

	mv = mem.MemoryValueFromFieldElement(yFiled)
	err = vm.Memory.WriteToAddress(&yAddr, &mv)
	if err != nil {
		return fmt.Errorf("write to y address %s: %w", yAddr, err)

	}
	return nil
}

type WideMul128 struct {
	lhs  hinter.ResOperander
	rhs  hinter.ResOperander
	high hinter.CellRefer
	low  hinter.CellRefer
}

func (hint *WideMul128) String() string {
	return "WideMul128"
}

func (hint *WideMul128) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	mask := &utils.Uint256Max128

	lhs, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %w", hint.lhs, err)
	}
	rhs, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %w", hint.rhs, err)
	}

	lhsFelt, err := lhs.FieldElement()
	if err != nil {
		return err
	}
	rhsFelt, err := rhs.FieldElement()
	if err != nil {
		return err
	}

	lhsU256 := uint256.Int(lhsFelt.Bits())
	rhsU256 := uint256.Int(rhsFelt.Bits())

	if lhsU256.Gt(mask) {
		return fmt.Errorf("lhs operand %s should be u128", lhsFelt)
	}
	if rhsU256.Gt(mask) {
		return fmt.Errorf("rhs operand %s should be u128", rhsFelt)
	}

	mul := lhsU256.Mul(&lhsU256, &rhsU256)

	bytes := mul.Bytes32()

	low := f.Element{}
	low.SetBytes(bytes[16:])

	high := f.Element{}
	high.SetBytes(bytes[:16])

	lowAddr, err := hint.low.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %w", err)
	}
	mvLow := mem.MemoryValueFromFieldElement(&low)
	err = vm.Memory.WriteToAddress(&lowAddr, &mvLow)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	highAddr, err := hint.high.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %w", err)
	}
	mvHigh := mem.MemoryValueFromFieldElement(&high)
	err = vm.Memory.WriteToAddress(&highAddr, &mvHigh)
	if err != nil {
		return fmt.Errorf("write cell: %w", err)
	}
	return nil
}

type DivMod struct {
	lhs       hinter.ResOperander
	rhs       hinter.ResOperander
	quotient  hinter.CellRefer
	remainder hinter.CellRefer
}

func (hint DivMod) String() string {
	return "DivMod"
}

func (hint DivMod) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {

	lhsVal, err := hint.lhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve lhs operand %s: %v", hint.lhs, err)
	}

	rhsVal, err := hint.rhs.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve rhs operand %s: %v", hint.rhs, err)
	}

	lhsFelt, err := lhsVal.FieldElement()
	if err != nil {
		return err
	}

	rhsFelt, err := rhsVal.FieldElement()
	if err != nil {
		return err
	}
	if rhsFelt.IsZero() {
		return fmt.Errorf("cannot be divided by zero, rhs: %v", rhsFelt)
	}

	lhsvalue := uint256.Int(lhsFelt.Bits())
	rhsvalue := uint256.Int(rhsFelt.Bits())

	// get quotient
	quo := uint256.Int{}
	quo.Div(&lhsvalue, &rhsvalue)

	quotient := f.Element{}
	quoVal := quo.Uint64()
	quotient.SetUint64(quoVal)

	quotientAddr, err := hint.quotient.Get(vm)
	if err != nil {
		return fmt.Errorf("get quotient cell: %v", err)
	}

	quotientVal := mem.MemoryValueFromFieldElement(&quotient)
	err = vm.Memory.WriteToAddress(&quotientAddr, &quotientVal)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	// get remainder: lhs - (rhs * quotient)
	temp := uint256.Int{}
	temp.Mul(&rhsvalue, &quo)

	rem := uint256.Int{}
	rem.Sub(&lhsvalue, &temp)

	remainder := f.Element{}
	remVal := rem.Uint64()
	remainder.SetUint64(remVal)

	remainderAddr, err := hint.remainder.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainder cell: %v", err)
	}

	remainderVal := mem.MemoryValueFromFieldElement(&remainder)
	err = vm.Memory.WriteToAddress(&remainderAddr, &remainderVal)
	if err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	return nil
}

type U256InvModN struct {
	b0           hinter.ResOperander
	b1           hinter.ResOperander
	n0           hinter.ResOperander
	n1           hinter.ResOperander
	g0_or_no_inv hinter.CellRefer
	g1_option    hinter.CellRefer
	s_or_r0      hinter.CellRefer
	s_or_r1      hinter.CellRefer
	t_or_k0      hinter.CellRefer
	t_or_k1      hinter.CellRefer
}

func (hint U256InvModN) String() string {
	return "U256InvModN"
}

func (hint U256InvModN) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	pow_2_128 := new(big.Int).Lsh(big.NewInt(1), 128)

	b0, err := hint.b0.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve b0 operand %s: %v", hint.b0, err)
	}
	b1, err := hint.b1.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve b1 operand %s: %v", hint.b1, err)
	}
	n0, err := hint.n0.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve n0 operand %s: %v", hint.n0, err)
	}
	n1, err := hint.n1.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve n1 operand %s: %v", hint.n1, err)
	}

	g0OrNoInvAddr, err := hint.g0_or_no_inv.Get(vm)
	if err != nil {
		return fmt.Errorf("get g0_or_no_inv address %s: %w", g0OrNoInvAddr, err)
	}
	g1OptionAddr, err := hint.g1_option.Get(vm)
	if err != nil {
		return fmt.Errorf("get g1_option address %s: %w", g1OptionAddr, err)
	}
	sOrR0Addr, err := hint.s_or_r0.Get(vm)
	if err != nil {
		return fmt.Errorf("get s_or_r0 address %s: %w", sOrR0Addr, err)
	}
	sOrR1Addr, err := hint.s_or_r1.Get(vm)
	if err != nil {
		return fmt.Errorf("get s_or_r1 address %s: %w", sOrR1Addr, err)
	}
	tOrK0Addr, err := hint.t_or_k0.Get(vm)
	if err != nil {
		return fmt.Errorf("get t_or_k0 address %s: %w", tOrK0Addr, err)
	}
	tOrK1Addr, err := hint.t_or_k1.Get(vm)
	if err != nil {
		return fmt.Errorf("get t_or_k1 address %s: %w", tOrK1Addr, err)
	}

	b0Felt, err := b0.FieldElement()
	if err != nil {
		return err
	}
	b1Felt, err := b1.FieldElement()
	if err != nil {
		return err
	}
	n0Felt, err := n0.FieldElement()
	if err != nil {
		return err
	}
	n1Felt, err := n1.FieldElement()
	if err != nil {
		return err
	}
	var b0BigInt big.Int
	b0Felt.BigInt(&b0BigInt)
	var b1BigInt big.Int
	b1Felt.BigInt(&b1BigInt)
	var n0BigInt big.Int
	n0Felt.BigInt(&n0BigInt)
	var n1BigInt big.Int
	n1Felt.BigInt(&n1BigInt)
	b := new(big.Int).Lsh(&b1BigInt, 128)
	b.Add(b, &b0BigInt)
	n := new(big.Int).Lsh(&n1BigInt, 128)
	n.Add(n, &n0BigInt)
	var x big.Int
	var y big.Int
	var g big.Int
	g.GCD(&x, &y, n, b)
	if n.Cmp(big.NewInt(1)) == 0 {
		mv := mem.MemoryValueFromFieldElement(b0Felt)
		err = vm.Memory.WriteToAddress(&sOrR0Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to s_or_r0 address %s: %w", sOrR0Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(b1Felt)
		err = vm.Memory.WriteToAddress(&sOrR1Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to s_or_r1 address %s: %w", sOrR1Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(&utils.FeltOne)
		err = vm.Memory.WriteToAddress(&tOrK0Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to t_or_k0 address %s: %w", tOrK0Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(&utils.FeltZero)
		err = vm.Memory.WriteToAddress(&tOrK1Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to t_or_k1 address %s: %w", tOrK1Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(&utils.FeltOne)
		err = vm.Memory.WriteToAddress(&g0OrNoInvAddr, &mv)
		if err != nil {
			return fmt.Errorf("write to g0_or_no_inv address %s: %w", g0OrNoInvAddr, err)
		}
		mv = mem.MemoryValueFromFieldElement(&utils.FeltZero)
		err = vm.Memory.WriteToAddress(&g1OptionAddr, &mv)
		if err != nil {
			return fmt.Errorf("write to g1_option address %s: %w", g1OptionAddr, err)
		}
	} else if g.Cmp(big.NewInt(1)) != 0 {
		if new(big.Int).Rem(&g, big.NewInt(2)) == big.NewInt(0) {
			g = *big.NewInt(2)
		}
		limb1 := new(big.Int).Div(new(big.Int).Div(b, &g), pow_2_128)
		limb0 := new(big.Int).Rem(new(big.Int).Div(b, &g), pow_2_128)
		mv := mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb0))
		err = vm.Memory.WriteToAddress(&sOrR0Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to s_or_r0 address %s: %w", sOrR0Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb1))
		err = vm.Memory.WriteToAddress(&sOrR1Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to s_or_r1 address %s: %w", sOrR1Addr, err)
		}
		limb1 = new(big.Int).Div(new(big.Int).Div(n, &g), pow_2_128)
		limb0 = new(big.Int).Rem(new(big.Int).Div(n, &g), pow_2_128)
		mv = mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb0))
		err = vm.Memory.WriteToAddress(&tOrK0Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to t_or_k0 address %s: %w", tOrK0Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb1))
		err = vm.Memory.WriteToAddress(&tOrK1Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to t_or_k1 address %s: %w", tOrK1Addr, err)
		}
		limb1 = new(big.Int).Div(&g, pow_2_128)
		limb0 = new(big.Int).Rem(&g, pow_2_128)
		mv = mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb0))
		err = vm.Memory.WriteToAddress(&g0OrNoInvAddr, &mv)
		if err != nil {
			return fmt.Errorf("write to g0_or_no_inv address %s: %w", g0OrNoInvAddr, err)
		}
		mv = mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb1))
		err = vm.Memory.WriteToAddress(&g1OptionAddr, &mv)
		if err != nil {
			return fmt.Errorf("write to g1_option address %s: %w", g1OptionAddr, err)
		}
	} else {
		y.Rem(&y, n)
		if y.Cmp(big.NewInt(0)) < 0 {
			y.Add(&y, n)
		}
		k := new(big.Int).Mul(&y, b)
		k.Sub(k, big.NewInt(1))
		k.Div(k, n)
		limb1 := new(big.Int).Div(&y, pow_2_128)
		limb0 := new(big.Int).Rem(&y, pow_2_128)
		mv := mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb0))
		err = vm.Memory.WriteToAddress(&sOrR0Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to s_or_r0 address %s: %w", sOrR0Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb1))
		err = vm.Memory.WriteToAddress(&sOrR1Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to s_or_r1 address %s: %w", sOrR1Addr, err)
		}
		limb1 = new(big.Int).Div(k, pow_2_128)
		limb0 = new(big.Int).Rem(k, pow_2_128)
		mv = mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb0))
		err = vm.Memory.WriteToAddress(&tOrK0Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to t_or_k0 address %s: %w", tOrK0Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(new(f.Element).SetBigInt(limb1))
		err = vm.Memory.WriteToAddress(&tOrK1Addr, &mv)
		if err != nil {
			return fmt.Errorf("write to t_or_k1 address %s: %w", tOrK1Addr, err)
		}
		mv = mem.MemoryValueFromFieldElement(&utils.FeltZero)
		err = vm.Memory.WriteToAddress(&g0OrNoInvAddr, &mv)
		if err != nil {
			return fmt.Errorf("write to g0_or_no_inv address %s: %w", g0OrNoInvAddr, err)
		}
		mv = mem.MemoryValueFromFieldElement(&utils.FeltOne)
		err = vm.Memory.WriteToAddress(&g1OptionAddr, &mv)
		if err != nil {
			return fmt.Errorf("write to g1_option address %s: %w", g1OptionAddr, err)
		}
	}

	// mv := mem.MemoryValueFromFieldElement(&resFelt)
	// err = vm.Memory.WriteToAddress(&dstAddr, &mv)
	// if err != nil {
	// 	return fmt.Errorf("write to dst address %s: %w", dstAddr, err)
	// }
	return nil
}

type Uint256DivMod struct {
	dividend0  hinter.ResOperander
	dividend1  hinter.ResOperander
	divisor0   hinter.ResOperander
	divisor1   hinter.ResOperander
	quotient0  hinter.CellRefer
	quotient1  hinter.CellRefer
	remainder0 hinter.CellRefer
	remainder1 hinter.CellRefer
}

func (hint Uint256DivMod) String() string {
	return "Uint256DivMod"
}

func (hint Uint256DivMod) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	dividend0, err := hint.dividend0.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve dividend0 operand %s: %v", hint.dividend0, err)
	}
	dividend0Felt, err := dividend0.FieldElement()
	if err != nil {
		return err
	}
	dividend1, err := hint.dividend1.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve dividend1 operand %s: %v", hint.dividend1, err)
	}
	dividend1Felt, err := dividend1.FieldElement()
	if err != nil {
		return err
	}
	divisor0, err := hint.divisor0.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve divisor0 operand %s: %v", hint.divisor0, err)
	}
	divisor0Felt, err := divisor0.FieldElement()
	if err != nil {
		return err
	}
	divisor1, err := hint.divisor1.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve divisor1 operand %s: %v", hint.divisor1, err)
	}
	divisor1Felt, err := divisor1.FieldElement()
	if err != nil {
		return err
	}

	var dividendBytes [32]byte
	dividend0Bytes := dividend0Felt.Bytes()
	dividend1Bytes := dividend1Felt.Bytes()
	copy(dividendBytes[:16], dividend1Bytes[16:])
	copy(dividendBytes[16:], dividend0Bytes[16:])

	dividend := &big.Int{}
	dividend.SetBytes(dividendBytes[:])

	var divisorBytes [32]byte
	divisor0Bytes := divisor0Felt.Bytes()
	divisor1Bytes := divisor1Felt.Bytes()
	copy(divisorBytes[:16], divisor1Bytes[16:])
	copy(divisorBytes[16:], divisor0Bytes[16:])

	divisor := &big.Int{}
	divisor.SetBytes(divisorBytes[:])
	if divisor.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("cannot be divided by zero, divisor: %v", divisor)
	}

	quotient, remainder := dividend.DivMod(dividend, divisor, &big.Int{})

	var quotientBytes [32]byte
	quotient.FillBytes(quotientBytes[:])
	quotientLimb1 := f.Element{}
	quotientLimb1.SetBytes(quotientBytes[:16])
	quotientLimb0 := f.Element{}
	quotientLimb0.SetBytes(quotientBytes[16:])

	var remainderBytes [32]byte
	remainder.FillBytes(remainderBytes[:])
	remainderLimb1 := f.Element{}
	remainderLimb1.SetBytes(remainderBytes[:16])
	remainderLimb0 := f.Element{}
	remainderLimb0.SetBytes(remainderBytes[16:])

	quotient0Addr, err := hint.quotient0.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	quotient0Val := mem.MemoryValueFromFieldElement(&quotientLimb0)

	if err = vm.Memory.WriteToAddress(&quotient0Addr, &quotient0Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	quotient1Addr, err := hint.quotient1.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	quotient1Val := mem.MemoryValueFromFieldElement(&quotientLimb1)
	if err = vm.Memory.WriteToAddress(&quotient1Addr, &quotient1Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	remainder0Addr, err := hint.remainder0.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	remainder0Val := mem.MemoryValueFromFieldElement(&remainderLimb0)
	if err = vm.Memory.WriteToAddress(&remainder0Addr, &remainder0Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}
	remainder1Addr, err := hint.remainder1.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	remainder1Val := mem.MemoryValueFromFieldElement(&remainderLimb1)
	if err = vm.Memory.WriteToAddress(&remainder1Addr, &remainder1Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}
	return nil
}

type DebugPrint struct {
	start hinter.ResOperander
	end   hinter.ResOperander
}

func (hint DebugPrint) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	start, err := hint.start.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve start operand %s: %v", hint.start, err)
	}

	startAddr, err := start.MemoryAddress()
	if err != nil {
		return fmt.Errorf("start memory address: %v", err)
	}

	end, err := hint.end.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve end operand %s: %v", hint.end, err)
	}
	endAddr, err := end.MemoryAddress()
	if err != nil {
		return fmt.Errorf("end memory address: %v", err)
	}

	if startAddr.Offset > endAddr.Offset {
		return fmt.Errorf("start cannot be greater than end")
	}

	current := startAddr.Offset
	for current < endAddr.Offset {
		v, err := vm.Memory.ReadFromAddress(&mem.MemoryAddress{
			SegmentIndex: startAddr.SegmentIndex,
			Offset:       current,
		})
		if err != nil {
			return err
		}

		field, _ := v.FieldElement()
		fmt.Printf("[DEBUG] %s\n", field.Text(16))
		current += 1
	}

	return nil
}

type SquareRoot struct {
	value hinter.ResOperander
	dst   hinter.CellRefer
}

func (hint *SquareRoot) String() string {
	return "SquareRoot"
}

func (hint *SquareRoot) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	value, err := hint.value.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve value operand %s: %w", hint.value, err)
	}

	valueFelt, err := value.FieldElement()
	if err != nil {
		return err
	}

	// Need to do this conversion to handle non-square values
	valueU256 := uint256.Int(valueFelt.Bits())
	valueU256.Sqrt(&valueU256)

	sqrt := f.Element{}
	sqrt.SetBytes(valueU256.Bytes())

	dstAddr, err := hint.dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %w", err)
	}

	dstVal := mem.MemoryValueFromFieldElement(&sqrt)
	err = vm.Memory.WriteToAddress(&dstAddr, &dstVal)
	if err != nil {
		return fmt.Errorf("write cell: %w", err)
	}
	return nil
}

type Uint256SquareRoot struct {
	valueLow                     hinter.ResOperander
	valueHigh                    hinter.ResOperander
	sqrt0                        hinter.CellRefer
	sqrt1                        hinter.CellRefer
	remainderLow                 hinter.CellRefer
	remainderHigh                hinter.CellRefer
	sqrtMul2MinusRemainderGeU128 hinter.CellRefer
}

func (hint Uint256SquareRoot) String() string {
	return "Uint256SquareRoot"
}

func (hint Uint256SquareRoot) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	valueLow, err := hint.valueLow.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve valueLow operand %s: %v", hint.valueLow, err)
	}

	valueHigh, err := hint.valueHigh.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve valueHigh operand %s: %v", hint.valueHigh, err)
	}

	valueLowFelt, err := valueLow.FieldElement()
	if err != nil {
		return err
	}

	valueHighFelt, err := valueHigh.FieldElement()
	if err != nil {
		return err
	}

	// value = {value_low} + {value_high} * 2**128
	valueLowU256 := uint256.Int(valueLowFelt.Bits())
	value := uint256.Int(valueHighFelt.Bits())
	value.Lsh(&value, 128)
	value.Add(&value, &valueLowU256)

	// root = math.isqrt(value)
	root := uint256.Int{}
	root.Sqrt(&value)

	// remainder = value - root ** 2
	root2 := uint256.Int{}
	root2.Mul(&root, &root)
	remainder := uint256.Int{}
	remainder.Sub(&value, &root2)

	// memory{sqrt0} = root & 0xFFFFFFFFFFFFFFFF
	// memory{sqrt1} = root >> 64
	mask64 := uint256.NewInt(0xFFFFFFFFFFFFFFFF)
	rootMasked := uint256.Int{}
	rootMasked.And(&root, mask64)
	rootShifted := root.Rsh(&root, 64)

	sqrt0 := f.Element{}
	sqrt0.SetBytes(rootMasked.Bytes())

	sqrt1 := f.Element{}
	sqrt1.SetBytes(rootShifted.Bytes())

	sqrt0Addr, err := hint.sqrt0.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrt0 cell: %v", err)
	}

	sqrt1Addr, err := hint.sqrt1.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrt1 cell: %v", err)
	}

	sqrt0Val := mem.MemoryValueFromFieldElement(&sqrt0)
	err = vm.Memory.WriteToAddress(&sqrt0Addr, &sqrt0Val)
	if err != nil {
		return fmt.Errorf("write sqrt0 cell: %v", err)
	}

	sqrt1Val := mem.MemoryValueFromFieldElement(&sqrt1)
	err = vm.Memory.WriteToAddress(&sqrt1Addr, &sqrt1Val)
	if err != nil {
		return fmt.Errorf("write sqrt1 cell: %v", err)
	}

	// memory{remainder_low} = remainder & 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
	// memory{remainder_high} = remainder >> 128
	mask128 := uint256.NewInt(0xFFFFFFFFFFFFFFFF)
	mask128.Lsh(mask128, 64)
	mask128.Or(mask128, mask64)
	remainderMasked := uint256.Int{}
	remainderMasked.And(&remainder, mask128)
	remainderLow := f.Element{}
	remainderLow.SetBytes(remainderMasked.Bytes())

	remainderShifted := uint256.Int{}
	remainderShifted.Rsh(&remainder, 128)
	remainderHigh := f.Element{}
	remainderHigh.SetBytes(remainderShifted.Bytes())

	remainderLowAddr, err := hint.remainderLow.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainderLow cell: %v", err)
	}

	remainderHighAddr, err := hint.remainderHigh.Get(vm)
	if err != nil {
		return fmt.Errorf("get remainderHigh cell: %v", err)
	}

	remainderLowVal := mem.MemoryValueFromFieldElement(&remainderLow)
	err = vm.Memory.WriteToAddress(&remainderLowAddr, &remainderLowVal)
	if err != nil {
		return fmt.Errorf("write remainderLow cell: %v", err)
	}

	remainderHighVal := mem.MemoryValueFromFieldElement(&remainderHigh)
	err = vm.Memory.WriteToAddress(&remainderHighAddr, &remainderHighVal)
	if err != nil {
		return fmt.Errorf("write remainderHigh cell: %v", err)
	}

	// memory{sqrt_mul_2_minus_remainder_ge_u128} = root * 2 - remainder >= 2**128
	rootMul2 := uint256.Int{}
	rootMul2.Lsh(&root, 1)
	lhs := uint256.Int{}
	lhs.Sub(&rootMul2, &remainder)

	rhs := uint256.NewInt(1)
	rhs.Lsh(rhs, 128)
	result := rhs.Gt(&lhs)
	result = !result

	sqrtMul2MinusRemainderGeU128 := f.Element{}
	if result {
		sqrtMul2MinusRemainderGeU128.SetOne()
	}

	sqrtMul2MinusRemainderGeU128Addr, err := hint.sqrtMul2MinusRemainderGeU128.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrtMul2MinusRemainderGeU128Addr cell: %v", err)
	}

	sqrtMul2MinusRemainderGeU128AddrVal := mem.MemoryValueFromFieldElement(&sqrtMul2MinusRemainderGeU128)
	err = vm.Memory.WriteToAddress(&sqrtMul2MinusRemainderGeU128Addr, &sqrtMul2MinusRemainderGeU128AddrVal)
	if err != nil {
		return fmt.Errorf("write sqrtMul2MinusRemainderGeU128Addr cell: %v", err)
	}

	return nil
}

//
// Dictionary Hints
//

type AllocFelt252Dict struct {
	SegmentArenaPtr hinter.ResOperander
}

func (hint *AllocFelt252Dict) String() string {
	return "AllocFelt252Dict"
}
func (hint *AllocFelt252Dict) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	hinter.InitializeDictionaryManager(ctx)

	arenaPtr, err := hinter.ResolveAsAddress(vm, hint.SegmentArenaPtr)
	if err != nil {
		return fmt.Errorf("resolve segment arena pointer: %w", err)
	}

	// find for the amount of initialized dicts
	initializedDictsOffset, overflow := utils.SafeOffset(arenaPtr.Offset, -2)
	if overflow {
		return fmt.Errorf("look for initialized dicts: overflow: %s - 2", arenaPtr)
	}
	initializedDictsFelt, err := vm.Memory.Read(arenaPtr.SegmentIndex, initializedDictsOffset)
	if err != nil {
		return fmt.Errorf("read initialized dicts: %w", err)
	}
	initializedDicts, err := initializedDictsFelt.Uint64()
	if err != nil {
		return fmt.Errorf("read initialized dicts: %w", err)
	}

	// find for the segment info pointer
	segmentInfoOffset, overflow := utils.SafeOffset(arenaPtr.Offset, -3)
	if overflow {
		return fmt.Errorf("look for segment info pointer: overflow: %s - 3", arenaPtr)
	}
	segmentInfoMv, err := vm.Memory.Read(arenaPtr.SegmentIndex, segmentInfoOffset)
	if err != nil {
		return fmt.Errorf("read segment info pointer: %w", err)
	}
	segmentInfoPtr, err := segmentInfoMv.MemoryAddress()
	if err != nil {
		return fmt.Errorf("expected pointer to segment info but got a felt: %w", err)
	}

	// with the segment info pointer and the number of initialized dictionaries we know
	// where to write the new dictionary
	newDictAddress := ctx.DictionaryManager.NewDictionary(vm)
	mv := mem.MemoryValueFromMemoryAddress(&newDictAddress)
	insertOffset := segmentInfoPtr.Offset + initializedDicts*3
	if err = vm.Memory.Write(segmentInfoPtr.SegmentIndex, insertOffset, &mv); err != nil {
		return fmt.Errorf("write new dictionary to segment info: %w", err)
	}
	return nil
}

type Felt252DictEntryInit struct {
	DictPtr hinter.ResOperander
	Key     hinter.ResOperander
}

func (hint Felt252DictEntryInit) String() string {
	return "Felt252DictEntryInit"
}

func (hint *Felt252DictEntryInit) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	dictPtr, err := hinter.ResolveAsAddress(vm, hint.DictPtr)
	if err != nil {
		return fmt.Errorf("resolve dictionary pointer: %w", err)
	}

	key, err := hinter.ResolveAsFelt(vm, hint.Key)
	if err != nil {
		return fmt.Errorf("resolve key: %w", err)
	}

	prevValue, err := ctx.DictionaryManager.At(dictPtr, key)
	if err != nil {
		return fmt.Errorf("get dictionary entry: %w", err)
	}
	if prevValue == nil {
		mv := mem.EmptyMemoryValueAsFelt()
		prevValue = &mv
		_ = ctx.DictionaryManager.Set(dictPtr, key, prevValue)
	}
	return vm.Memory.Write(dictPtr.SegmentIndex, dictPtr.Offset+1, prevValue)
}

type Felt252DictEntryUpdate struct {
	DictPtr hinter.ResOperander
	Value   hinter.ResOperander
}

func (hint Felt252DictEntryUpdate) String() string {
	return "Felt252DictEntryUpdate"
}

func (hint *Felt252DictEntryUpdate) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	dictPtr, err := hinter.ResolveAsAddress(vm, hint.DictPtr)
	if err != nil {
		return fmt.Errorf("resolve dictionary pointer: %w", err)
	}

	keyPtr, err := dictPtr.AddOffset(-3)
	if err != nil {
		return fmt.Errorf("get key pointer: %w", err)
	}
	keyMv, err := vm.Memory.ReadFromAddress(&keyPtr)
	if err != nil {
		return fmt.Errorf("read key pointer: %w", err)
	}
	key, err := keyMv.FieldElement()
	if err != nil {
		return fmt.Errorf("expected key to be a field element: %w", err)
	}

	value, err := hint.Value.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve value: %w", err)
	}

	return ctx.DictionaryManager.Set(dictPtr, key, &value)
}

type GetSegmentArenaIndex struct {
	DictIndex  hinter.CellRefer
	DictEndPtr hinter.ResOperander
}

func (hint *GetSegmentArenaIndex) String() string {
	return "GetSegmentArenaIndex"
}

func (hint *GetSegmentArenaIndex) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	dictIndex, err := hint.DictIndex.Get(vm)
	if err != nil {
		return fmt.Errorf("get dict index: %w", err)
	}

	dictEndPtr, err := hinter.ResolveAsAddress(vm, hint.DictEndPtr)
	if err != nil {
		return fmt.Errorf("resolve dict end pointer: %w", err)
	}

	dict, err := ctx.DictionaryManager.GetDictionary(dictEndPtr)
	if err != nil {
		return fmt.Errorf("get dictionary: %w", err)
	}

	initNum := mem.MemoryValueFromUint(dict.InitNumber())
	return vm.Memory.WriteToAddress(&dictIndex, &initNum)
}

//
// Squashed Dictionary Hints
//

type InitSquashData struct {
	FirstKey     hinter.CellRefer
	BigKeys      hinter.CellRefer
	DictAccesses hinter.ResOperander
	NumAccesses  hinter.ResOperander
}

func (hint *InitSquashData) String() string {
	return "InitSquashData"
}

func (hint *InitSquashData) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	// todo(rodro): Don't know if it could be called multiple times, or
	err := hinter.InitializeSquashedDictionaryManager(ctx)
	if err != nil {
		return err
	}

	dictAccessPtr, err := hinter.ResolveAsAddress(vm, hint.DictAccesses)
	if err != nil {
		return fmt.Errorf("resolve dict access: %w", err)
	}

	numAccess, err := hinter.ResolveAsUint64(vm, hint.NumAccesses)
	if err != nil {
		return fmt.Errorf("resolve num access: %w", err)
	}

	const dictAccessSize = 3
	for i := uint64(0); i < numAccess; i++ {
		keyPtr := mem.MemoryAddress{
			SegmentIndex: dictAccessPtr.SegmentIndex,
			Offset:       dictAccessPtr.Offset + i*dictAccessSize,
		}
		key, err := vm.Memory.ReadFromAddressAsElement(&keyPtr)
		if err != nil {
			return fmt.Errorf("reading key at %s: %w", keyPtr, err)
		}

		ctx.SquashedDictionaryManager.Insert(&key, i)
	}
	for key, val := range ctx.SquashedDictionaryManager.KeyToIndices {
		// reverse each indice access list per key
		utils.Reverse(val)
		// store each key
		ctx.SquashedDictionaryManager.Keys = append(ctx.SquashedDictionaryManager.Keys, key)
	}

	// sort the keys in descending order
	sort.Slice(ctx.SquashedDictionaryManager.Keys, func(i, j int) bool {
		return ctx.SquashedDictionaryManager.Keys[i].Cmp(&ctx.SquashedDictionaryManager.Keys[j]) < 0
	})

	// if the first key is bigger than 2^128, signal it
	bigKeysAddr, err := hint.BigKeys.Get(vm)
	if err != nil {
		return fmt.Errorf("get big keys address: %w", err)
	}
	biggestKey := ctx.SquashedDictionaryManager.Keys[0]
	cmpRes := mem.MemoryValueFromUint[uint64](0)
	if biggestKey.Cmp(&utils.FeltMax128) > 0 {
		cmpRes = mem.MemoryValueFromUint[uint64](1)
	}
	err = vm.Memory.WriteToAddress(&bigKeysAddr, &cmpRes)
	if err != nil {
		return fmt.Errorf("write big keys address: %w", err)
	}

	// store the left most, smaller key
	firstKeyAddr, err := hint.FirstKey.Get(vm)
	if err != nil {
		return fmt.Errorf("get first key address: %w", err)
	}
	firstKey, err := ctx.SquashedDictionaryManager.LastKey()
	if err != nil {
		return fmt.Errorf("get first key: %w", err)
	}

	mv := mem.MemoryValueFromFieldElement(&firstKey)
	return vm.Memory.WriteToAddress(&firstKeyAddr, &mv)
}

type GetCurrentAccessIndex struct {
	RangeCheckPtr hinter.ResOperander
}

func (hint *GetCurrentAccessIndex) String() string {
	return "GetCurrentAccessIndex"
}

func (hint *GetCurrentAccessIndex) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	rangeCheckPtr, err := hinter.ResolveAsAddress(vm, hint.RangeCheckPtr)
	if err != nil {
		return fmt.Errorf("resolve range check pointer: %w", err)
	}

	lastIndex64, err := ctx.SquashedDictionaryManager.LastIndex()
	if err != nil {
		return fmt.Errorf("get last index: %w", err)
	}

	lastIndex := f.NewElement(lastIndex64)
	mv := mem.MemoryValueFromFieldElement(&lastIndex)

	return vm.Memory.WriteToAddress(rangeCheckPtr, &mv)
}

type ShouldSkipSquashLoop struct {
	ShouldSkipLoop hinter.CellRefer
}

func (hint *ShouldSkipSquashLoop) String() string {
	return "ShouldSkipSquashLoop"
}

func (hint *ShouldSkipSquashLoop) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	shouldSkipLoopAddr, err := hint.ShouldSkipLoop.Get(vm)
	if err != nil {
		return fmt.Errorf("get should skip loop address: %w", err)
	}

	shouldSkipLoop := &f.Element{}
	shouldSkipLoop.SetOne()
	if lastIndices, err := ctx.SquashedDictionaryManager.LastIndices(); err == nil && len(lastIndices) > 1 {
		shouldSkipLoop.SetZero()
	} else if err != nil {
		return fmt.Errorf("get last indices: %w", err)
	}

	mv := mem.MemoryValueFromFieldElement(shouldSkipLoop)
	return vm.Memory.WriteToAddress(&shouldSkipLoopAddr, &mv)
}

type GetCurrentAccessDelta struct {
	IndexDeltaMinusOne hinter.CellRefer
}

func (hint *GetCurrentAccessDelta) String() string {
	return "GetCurrentAccessDelta"
}

func (hint *GetCurrentAccessDelta) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	indexDeltaPtr, err := hint.IndexDeltaMinusOne.Get(vm)
	if err != nil {
		return fmt.Errorf("get index delta address: %w", err)
	}

	previousKeyIndex, err := ctx.SquashedDictionaryManager.PopIndex()
	if err != nil {
		return fmt.Errorf("pop index: %w", err)
	}

	currentKeyIndex, err := ctx.SquashedDictionaryManager.LastIndex()
	if err != nil {
		return fmt.Errorf("get last index: %w", err)
	}

	// todo(rodro): could previousKeyIndex be bigger than currentKeyIndex?
	indexDeltaMinusOne := currentKeyIndex - previousKeyIndex - 1
	mv := mem.MemoryValueFromUint(indexDeltaMinusOne)

	return vm.Memory.WriteToAddress(&indexDeltaPtr, &mv)
}

type ShouldContinueSquashLoop struct {
	ShouldContinue hinter.CellRefer
}

func (hint *ShouldContinueSquashLoop) String() string {
	return "ShouldContinueSquashLoop"
}

func (hint *ShouldContinueSquashLoop) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	shouldContinuePtr, err := hint.ShouldContinue.Get(vm)
	if err != nil {
		return fmt.Errorf("get should continue address: %w", err)
	}

	var shouldContinueLoop f.Element
	if lastIndices, err := ctx.SquashedDictionaryManager.LastIndices(); err == nil && len(lastIndices) <= 1 {
		shouldContinueLoop.SetOne()
	} else if err != nil {
		return fmt.Errorf("get last indices: %w", err)
	}

	mv := mem.MemoryValueFromFieldElement(&shouldContinueLoop)
	return vm.Memory.WriteToAddress(&shouldContinuePtr, &mv)
}

type GetNextDictKey struct {
	NextKey hinter.CellRefer
}

func (hint *GetNextDictKey) String() string {
	return "GetNextDictKey"
}

func (hint *GetNextDictKey) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	nextKeyAddr, err := hint.NextKey.Get(vm)
	if err != nil {
		return fmt.Errorf("get next key address: %w", err)
	}

	nextKey, err := ctx.SquashedDictionaryManager.PopKey()
	if err != nil {
		return fmt.Errorf("pop key: %w", err)
	}

	mv := mem.MemoryValueFromFieldElement(&nextKey)
	return vm.Memory.WriteToAddress(&nextKeyAddr, &mv)
}

type Uint512DivModByUint256 struct {
	dividend0  hinter.ResOperander
	dividend1  hinter.ResOperander
	dividend2  hinter.ResOperander
	dividend3  hinter.ResOperander
	divisor0   hinter.ResOperander
	divisor1   hinter.ResOperander
	quotient0  hinter.CellRefer
	quotient1  hinter.CellRefer
	quotient2  hinter.CellRefer
	quotient3  hinter.CellRefer
	remainder0 hinter.CellRefer
	remainder1 hinter.CellRefer
}

func (hint Uint512DivModByUint256) String() string {
	return "Uint512DivModByUint256"
}

func (hint Uint512DivModByUint256) Execute(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
	dividend0, err := hint.dividend0.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve dividend0 operand %s: %v", hint.dividend0, err)
	}
	dividend0Felt, err := dividend0.FieldElement()
	if err != nil {
		return err
	}
	dividend1, err := hint.dividend1.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve dividend1 operand %s: %v", hint.dividend1, err)
	}
	dividend1Felt, err := dividend1.FieldElement()
	if err != nil {
		return err
	}
	dividend2, err := hint.dividend2.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve dividend2 operand %s: %v", hint.dividend2, err)
	}
	dividend2Felt, err := dividend2.FieldElement()
	if err != nil {
		return err
	}
	dividend3, err := hint.dividend3.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve dividend3 operand %s: %v", hint.dividend3, err)
	}
	dividend3Felt, err := dividend3.FieldElement()
	if err != nil {
		return err
	}

	divisor0, err := hint.divisor0.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve divisor0 operand %s: %v", hint.divisor0, err)
	}
	divisor0Felt, err := divisor0.FieldElement()
	if err != nil {
		return err
	}
	divisor1, err := hint.divisor1.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve divisor1 operand %s: %v", hint.divisor1, err)
	}
	divisor1Felt, err := divisor1.FieldElement()
	if err != nil {
		return err
	}

	var dividendBytes [64]byte
	dividend0Bytes := dividend0Felt.Bytes()
	dividend1Bytes := dividend1Felt.Bytes()
	dividend2Bytes := dividend2Felt.Bytes()
	dividend3Bytes := dividend3Felt.Bytes()
	copy(dividendBytes[:16], dividend3Bytes[16:])
	copy(dividendBytes[16:32], dividend2Bytes[16:])
	copy(dividendBytes[32:48], dividend1Bytes[16:])
	copy(dividendBytes[48:], dividend0Bytes[16:])
	// TODO: check for possible use of uint512 module.
	dividend := &big.Int{}
	dividend.SetBytes(dividendBytes[:])

	var divisorBytes [32]byte
	divisor0Bytes := divisor0Felt.Bytes()
	divisor1Bytes := divisor1Felt.Bytes()
	copy(divisorBytes[:16], divisor1Bytes[16:])
	copy(divisorBytes[16:], divisor0Bytes[16:])
	// TODO: check for possible use of uint512 module.
	divisor := &big.Int{}
	divisor.SetBytes(divisorBytes[:])
	if divisor.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("division by zero")
	}

	quotient, rem := dividend.DivMod(dividend, divisor, &big.Int{})

	var qBytes [64]byte
	quotient.FillBytes(qBytes[:])
	qlimb3 := f.Element{}
	qlimb3.SetBytes(qBytes[:16])
	qlimb2 := f.Element{}
	qlimb2.SetBytes(qBytes[16:32])
	qlimb1 := f.Element{}
	qlimb1.SetBytes(qBytes[32:48])
	qlimb0 := f.Element{}
	qlimb0.SetBytes(qBytes[48:])

	var rBytes [32]byte
	rem.FillBytes(rBytes[:])
	rlimb1 := f.Element{}
	rlimb1.SetBytes(rBytes[:16])
	rlimb0 := f.Element{}
	rlimb0.SetBytes(rBytes[16:])

	quotient0Addr, err := hint.quotient0.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	quotient0Val := mem.MemoryValueFromFieldElement(&qlimb0)

	if err = vm.Memory.WriteToAddress(&quotient0Addr, &quotient0Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	quotient1Addr, err := hint.quotient1.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	quotient1Val := mem.MemoryValueFromFieldElement(&qlimb1)
	if err = vm.Memory.WriteToAddress(&quotient1Addr, &quotient1Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	quotient2Addr, err := hint.quotient2.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	quotient2Val := mem.MemoryValueFromFieldElement(&qlimb2)
	if err = vm.Memory.WriteToAddress(&quotient2Addr, &quotient2Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	quotient3Addr, err := hint.quotient3.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	quotient3Val := mem.MemoryValueFromFieldElement(&qlimb3)
	if err = vm.Memory.WriteToAddress(&quotient3Addr, &quotient3Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	remainder0Addr, err := hint.remainder0.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	remainder0Val := mem.MemoryValueFromFieldElement(&rlimb0)
	if err = vm.Memory.WriteToAddress(&remainder0Addr, &remainder0Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	remainder1Addr, err := hint.remainder1.Get(vm)
	if err != nil {
		return fmt.Errorf("get destination cell: %v", err)
	}
	remainder1Val := mem.MemoryValueFromFieldElement(&rlimb1)
	if err = vm.Memory.WriteToAddress(&remainder1Addr, &remainder1Val); err != nil {
		return fmt.Errorf("write cell: %v", err)
	}

	return nil
}

type AllocConstantSize struct {
	Size hinter.ResOperander
	Dst  hinter.CellRefer
}

func (hint *AllocConstantSize) String() string {
	return "AllocConstantSize"
}

func (hint *AllocConstantSize) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	size, err := hinter.ResolveAsUint64(vm, hint.Size)
	if err != nil {
		return fmt.Errorf("size to big: %w", err)
	}

	if ctx.ConstantSizeSegment.Equal(&mem.UnknownAddress) {
		ctx.ConstantSizeSegment = vm.Memory.AllocateEmptySegment()
	}

	dst, err := hint.Dst.Get(vm)
	if err != nil {
		return fmt.Errorf("get dst %w", err)
	}

	val := mem.MemoryValueFromMemoryAddress(&ctx.ConstantSizeSegment)
	if err = vm.Memory.WriteToAddress(&dst, &val); err != nil {
		return err
	}

	// todo(rodro): Is possible for this to overflow
	ctx.ConstantSizeSegment.Offset += size
	return nil
}

type AssertLeFindSmallArc struct {
	A             hinter.ResOperander
	B             hinter.ResOperander
	RangeCheckPtr hinter.ResOperander
}

func (hint *AssertLeFindSmallArc) String() string {
	return "AssertLeFindSmallArc"
}

func (hint *AssertLeFindSmallArc) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	primeOver3High := uint256.Int{6148914691236517206, 192153584101141168, 0, 0}
	primeOver2High := uint256.Int{9223372036854775809, 288230376151711752, 0, 0}

	a, err := hint.A.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve a operand %s: %w", hint.A, err)
	}

	b, err := hint.B.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve b operand %s: %w", hint.B, err)
	}

	aFelt, err := a.FieldElement()
	if err != nil {
		return err
	}

	bFelt, err := b.FieldElement()
	if err != nil {
		return err
	}

	thirdLength := f.Element{32, 0, 0, 544} // -1 field element
	thirdLength.Sub(&thirdLength, bFelt)

	// Array of pairs (2-tuple)
	lengthsAndIndices := []struct {
		Value    f.Element
		Position int
	}{
		{*aFelt, 0},
		{*bFelt.Sub(bFelt, aFelt), 1},
		{thirdLength, 2},
	}

	// Sort
	sort.Slice(lengthsAndIndices, func(i, j int) bool {
		// lengthsAndIndices[i].Value < lengthsAndIndices[j].Value
		return lengthsAndIndices[i].Value.Cmp(&lengthsAndIndices[j].Value) == -1
	})

	// Exclude the largest arc after sorting
	err = ctx.ScopeManager.AssignVariable("excluded", lengthsAndIndices[2].Position)
	if err != nil {
		return err
	}

	rangeCheckPtrMemAddr, err := hinter.ResolveAsAddress(vm, hint.RangeCheckPtr)
	if err != nil {
		return fmt.Errorf("resolve range check pointer: %w", err)
	}

	// Find the quotient and modulo of the first element's value of the sorted array
	// w.r.t. primeOver3High
	quotient := uint256.Int(lengthsAndIndices[0].Value.Bits())
	remainder := uint256.Int{}
	quotient.DivMod(&quotient, &primeOver3High, &remainder)

	remainderFelt := f.Element{}
	remainderFelt.SetBytes(remainder.Bytes())

	quotientFelt := f.Element{}
	quotientFelt.SetBytes(quotient.Bytes())

	// Store remainder 1
	writeValue := mem.MemoryValueFromFieldElement(&remainderFelt)
	err = vm.Memory.WriteToAddress(rangeCheckPtrMemAddr, &writeValue)
	if err != nil {
		return fmt.Errorf("write first remainder: %w", err)
	}

	// Store quotient 1
	rangeCheckPtrMemAddr.Offset += 1
	writeValue = mem.MemoryValueFromFieldElement(&quotientFelt)
	err = vm.Memory.WriteToAddress(rangeCheckPtrMemAddr, &writeValue)
	if err != nil {
		return fmt.Errorf("write first quotient: %w", err)
	}

	// Find the quotient and modulo of the second element' value of the sorted array
	// w.r.t. primeOver2High
	quotient = uint256.Int(lengthsAndIndices[1].Value.Bits())
	remainder = uint256.Int{}
	quotient.DivMod(&quotient, &primeOver2High, &remainder)

	remainderFelt.SetBytes(remainder.Bytes())
	quotientFelt.SetBytes(quotient.Bytes())

	// Store remainder 2
	rangeCheckPtrMemAddr.Offset += 1
	writeValue = mem.MemoryValueFromFieldElement(&remainderFelt)
	err = vm.Memory.WriteToAddress(rangeCheckPtrMemAddr, &writeValue)
	if err != nil {
		return fmt.Errorf("write second remainder: %w", err)
	}

	// Store quotient 2
	rangeCheckPtrMemAddr.Offset += 1
	writeValue = mem.MemoryValueFromFieldElement(&quotientFelt)
	err = vm.Memory.WriteToAddress(rangeCheckPtrMemAddr, &writeValue)
	if err != nil {
		return fmt.Errorf("write second quotient: %w", err)
	}
	return nil
}

type AssertLeIsFirstArcExcluded struct {
	SkipExcludeAFlag hinter.CellRefer
}

func (hint *AssertLeIsFirstArcExcluded) String() string {
	return "AssertLeIsFirstArcExcluded"
}

func (hint *AssertLeIsFirstArcExcluded) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	addr, err := hint.SkipExcludeAFlag.Get(vm)
	if err != nil {
		return fmt.Errorf("get skipExcludeAFlag addr: %v", err)
	}

	var writeValue mem.MemoryValue
	excluded, err := ctx.ScopeManager.GetVariableValue("excluded")
	if err != nil {
		return err
	}
	if excluded != 0 {
		writeValue = mem.MemoryValueFromInt(1)
	} else {
		writeValue = mem.MemoryValueFromInt(0)
	}

	return vm.Memory.WriteToAddress(&addr, &writeValue)
}

type AssertLeIsSecondArcExcluded struct {
	SkipExcludeBMinusA hinter.CellRefer
}

func (hint *AssertLeIsSecondArcExcluded) String() string {
	return "AssertLeIsSecondArcExcluded"
}

func (hint *AssertLeIsSecondArcExcluded) Execute(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
	addr, err := hint.SkipExcludeBMinusA.Get(vm)
	if err != nil {
		return fmt.Errorf("get skipExcludeBMinusA addr: %v", err)
	}

	var writeValue mem.MemoryValue
	excluded, err := ctx.ScopeManager.GetVariableValue("excluded")
	if err != nil {
		return err
	}
	if excluded != 1 {
		writeValue = mem.MemoryValueFromInt(1)
	} else {
		writeValue = mem.MemoryValueFromInt(0)
	}

	return vm.Memory.WriteToAddress(&addr, &writeValue)
}

type RandomEcPoint struct {
	x hinter.CellRefer
	y hinter.CellRefer
}

func (hint *RandomEcPoint) String() string {
	return "RandomEc"
}

func (hint *RandomEcPoint) Execute(vm *VM.VirtualMachine) error {
	// Keep sampling a random field element `X` until `X^3 + X + beta` is a quadratic residue.

	// Starkware's elliptic curve Beta value https://docs.starkware.co/starkex/crypto/stark-curve.html
	betaFelt := f.Element{3863487492851900874, 7432612994240712710, 12360725113329547591, 88155977965380735}

	var randomX, randomYSquared f.Element
	rand := u.DefaultRandGenerator()
	for {
		randomX = u.RandomFeltElement(rand)
		randomYSquared = f.Element{}
		randomYSquared.Square(&randomX)
		randomYSquared.Mul(&randomYSquared, &randomX)
		randomYSquared.Add(&randomYSquared, &randomX)
		randomYSquared.Add(&randomYSquared, &betaFelt)
		// Legendre == 1 -> Quadratic residue
		// Legendre == -1 -> Quadratic non-residue
		// Legendre == 0 -> Zero
		// https://en.wikipedia.org/wiki/Legendre_symbol
		if randomYSquared.Legendre() == 1 {
			break
		}
	}

	xVal := mem.MemoryValueFromFieldElement(&randomX)
	yVal := mem.MemoryValueFromFieldElement(randomYSquared.Square(&randomYSquared))

	xAddr, err := hint.x.Get(vm)
	if err != nil {
		return fmt.Errorf("get register x %s: %w", hint.x, err)
	}

	err = vm.Memory.WriteToAddress(&xAddr, &xVal)
	if err != nil {
		return fmt.Errorf("write to address x %s: %w", xVal, err)
	}

	yAddr, err := hint.y.Get(vm)
	if err != nil {
		return fmt.Errorf("get register y %s: %w", hint.y, err)
	}

	err = vm.Memory.WriteToAddress(&yAddr, &yVal)
	if err != nil {
		return fmt.Errorf("write to address y %s: %w", yVal, err)
	}

	return nil
}

type FieldSqrt struct {
	val  hinter.ResOperander
	sqrt hinter.CellRefer
}

func (hint *FieldSqrt) String() string {
	return "FieldSqrt"
}

func (hint *FieldSqrt) Execute(vm *VM.VirtualMachine) error {
	val, err := hint.val.Resolve(vm)
	if err != nil {
		return fmt.Errorf("resolve val operand %s: %v", hint.val, err)
	}

	valFelt, err := val.FieldElement()
	if err != nil {
		return err
	}

	threeFelt := f.Element{}
	threeFelt.SetUint64(3)

	var res f.Element
	// Legendre == 1 -> Quadratic residue
	// Legendre == -1 -> Quadratic non-residue
	// Legendre == 0 -> Zero
	// https://en.wikipedia.org/wiki/Legendre_symbol
	if valFelt.Legendre() == 1 {
		res = *valFelt
	} else {
		res = *valFelt.Mul(valFelt, &threeFelt)
	}

	res.Sqrt(&res)

	sqrtVal := mem.MemoryValueFromFieldElement(&res)

	sqrtAddr, err := hint.sqrt.Get(vm)
	if err != nil {
		return fmt.Errorf("get sqrt address: %v", err)
	}

	return vm.Memory.WriteToAddress(&sqrtAddr, &sqrtVal)
}
