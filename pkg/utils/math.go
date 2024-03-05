package utils

import (
	"math/big"
	"math/bits"

	"golang.org/x/exp/constraints"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Takes a uint64 and an int16 and outputs their addition as well
// as the ocurrence of an overflow or underflow.
//
// This is a constant-time version of the following function:
//
//	func SafeOffset(x uint64, y int16) (res uint64, isOverflow bool) {
//		res = x + uint64(y)
//		if y < 0 {
//			isOverflow = res >= x
//		} else {
//			isOverflow = res < x
//		}
//		return
//	}
//
// This shows better results because the final bytecode
// doesn't contain any conditional jump instructions
// making it easier for a processor to pipeline the function.
func SafeOffset(x uint64, y int16) (res uint64, isOverflow bool) {
	enlargedY := uint64(y)
	// I'll leave proving that this is correct as an exercise for the reader :)
	res = x + enlargedY
	// Why does this work?
	// Let's proceed by cases on the most significant bit of (MSB(x)).
	// If MSB(x) == 1 and y < 0 (MSB(y) == 1) then overflow doesn't happen.
	// Let's consider the case y >= 0 (MSB(y) == 0).
	// In that case we can only wrap up by going to the begining of uint64 range making the MSB(res) = 0.
	// This is the second disjunct of the disjunctive formula.
	//
	// In the same fashion the case MSB(x) == 0 and MSB(y) == 1 is reasoned about.
	//
	// Finally, we boil everything down to MSBs by rotating and anding with ...000001.
	isOverflow = bits.RotateLeft64((^x&enlargedY&res)|(x & ^enlargedY & ^res), 1)&0x1 != 0
	return
}

// Given a number returns its closest power of two bigger than the number
func NextPowerOfTwo(n uint64) uint64 {
	// it is already a power of 2
	if (n & (n - 1)) == 0 {
		return n
	}

	higherBit := 64 - bits.LeadingZeros64(n)
	return 1 << higherBit
}

func Max[T constraints.Integer](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// FeltLt implements `a < b` felt comparison.
func FeltLt(a, b *fp.Element) bool {
	return a.Cmp(b) == -1
}

// FeltLe implements `a <= b` felt comparison.
func FeltLe(a, b *fp.Element) bool {
	// a is less or equal than b if it's not greater than b.
	return a.Cmp(b) != 1
}

func FeltIsPositive(felt *fp.Element) bool {
	// range_check_builtin.bound is utils.FeltMax128 (1 << 128).
	return FeltLt(felt, &FeltMax128)
}

// FeltMod implements `a % b` operation.
func FeltMod(a, b *fp.Element) fp.Element {
	// TODO: implement it in a better way, without bigint?

	var result fp.Element

	var tmpResult big.Int
	var tmpA big.Int
	var tmpB big.Int

	a.BigInt(&tmpA)
	b.BigInt(&tmpB)
	tmpResult.Mod(&tmpA, &tmpB)

	result.SetBigInt(&tmpResult)
	return result
}

func FeltDivRem(a, b *fp.Element) (div fp.Element, rem fp.Element) {
	// It would be possible to compute the mod (rem) as `a - div*b`,
	// but since felt div would yield a different result than bigint
	// arithmetics, we can't use that trick here.
	// divmod function used in Python cairovm does a non-felt divmod.
	// Therefore, 450326666 / 136310839 is expected to have a result of 3,
	// not 834010808316774569532950779803492285717614100391395442358316910417277897363.

	var tmpA big.Int
	var tmpB big.Int
	var tmpDiv big.Int
	var tmpRem big.Int
	a.BigInt(&tmpA)
	b.BigInt(&tmpB)
	tmpDiv.DivMod(&tmpA, &tmpB, &tmpRem)

	div.SetBigInt(&tmpDiv)
	rem.SetBigInt(&tmpRem)

	return div, rem
}
