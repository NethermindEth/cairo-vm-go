package safemath

import (
	"math/bits"
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
		return uint64(n)
	}

	higherBit := 64 - bits.LeadingZeros64(n)
	return 1 << higherBit
}
