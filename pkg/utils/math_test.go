package utils

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestOffsetNeg(t *testing.T) {
	res, isOverflow := SafeOffset(1215, -3)
	assert.Equal(t, uint64(1212), res)
	assert.False(t, isOverflow)
}

func TestOffsetPos(t *testing.T) {
	res, isOverflow := SafeOffset(7, 11)
	assert.Equal(t, uint64(18), res)
	assert.False(t, isOverflow)
}

func TestOffsetLeftOverflow(t *testing.T) {
	_, isOverflow := SafeOffset(4, -10)
	assert.True(t, isOverflow)
}

func TestOffsetRightOverflow(t *testing.T) {
	_, isOverflow := SafeOffset(^uint64(0), 1)
	assert.True(t, isOverflow)
}

func TestOffsetRightNoOverflow(t *testing.T) {
	res, isOverflow := SafeOffset(^uint64(0), -12)
	assert.Equal(t, uint64(18446744073709551603), res)
	assert.False(t, isOverflow)
}

func TestFeltDivRem(t *testing.T) {
	type testCase struct {
		a   fp.Element
		b   fp.Element
		div fp.Element
		rem fp.Element
	}
	tests := []testCase{
		{fp.NewElement(0), fp.NewElement(1), fp.NewElement(0), fp.NewElement(0)},
		{fp.NewElement(10), fp.NewElement(2), fp.NewElement(5), fp.NewElement(0)},
		{fp.NewElement(2), fp.NewElement(10), fp.NewElement(0), fp.NewElement(2)},
		{fp.NewElement(10), fp.NewElement(9), fp.NewElement(1), fp.NewElement(1)},
		{fp.NewElement(9), fp.NewElement(10), fp.NewElement(0), fp.NewElement(9)},
		{fp.NewElement(102495), fp.NewElement(2), fp.NewElement(51247), fp.NewElement(1)},
		{fp.NewElement(102495), fp.NewElement(23), fp.NewElement(4456), fp.NewElement(7)},
		{fp.NewElement(102495), fp.NewElement(5), fp.NewElement(20499), fp.NewElement(0)},
		{fp.NewElement(102495), fp.NewElement(102495), fp.NewElement(1), fp.NewElement(0)},
		{fp.NewElement(102495), fp.NewElement(102495 / 5), fp.NewElement(5), fp.NewElement(0)},
	}

	for i, test := range tests {
		haveDiv, haveRem := FeltDivRem(&test.a, &test.b)

		if !test.div.Equal(&haveDiv) || !test.rem.Equal(&haveRem) {
			t.Fatalf("test[%d]: %v divmod %v results mismatched:\nhave: %v, %v\nwant: %v, %v",
				i, &test.a, &test.b, &haveDiv, &haveRem, test.div, test.rem)
		}
	}
}
