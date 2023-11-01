package utils

import (
	"testing"

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
