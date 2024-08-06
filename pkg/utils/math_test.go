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

func TestRightRot(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint32
		n        uint32
		expected uint32
	}{
		{
			name:     "Rotate 0 bits",
			value:    0x12345678,
			n:        0,
			expected: 0x12345678,
		},
		{
			name:     "Rotate 1 bit",
			value:    0x12345678,
			n:        1,
			expected: 0x091A2B3C,
		},
		{
			name:     "Rotate 4 bits",
			value:    0x12345678,
			n:        4,
			expected: 0x81234567,
		},
		{
			name:     "Rotate 8 bits",
			value:    0x12345678,
			n:        8,
			expected: 0x78123456,
		},
		{
			name:     "Rotate 16 bits",
			value:    0x12345678,
			n:        16,
			expected: 0x56781234,
		},
		{
			name:     "Rotate 31 bits",
			value:    0x12345678,
			n:        31,
			expected: 0x2468ACF0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := RightRot(tc.value, tc.n)
			if result != tc.expected {
				t.Errorf("Expected %08X, got %08X", tc.expected, result)
			}
		})
	}
}
