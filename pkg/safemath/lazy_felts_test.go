package safemath

import (
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
)

func TestAdd2Uints(t *testing.T) {
	a, b := new(LazyFelt).SetUval(10), new(LazyFelt).SetUval(20)
	c := new(LazyFelt)

	c = c.Add(a, b)

	assert.True(t, c.IsUint64())
	assert.Equal(t, uint64(30), c.Uint64())
}

func TestAdd2UintsOverflow1(t *testing.T) {
	a, b := new(LazyFelt).SetUval(^uint64(0)), new(LazyFelt).SetUval(20)
	c := new(LazyFelt)

	c = c.Add(a, b)

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Add(new(f.Element).SetUint64(^uint64(0)), new(f.Element).SetUint64(20))
	assert.False(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestAdd2UintsOverflow2(t *testing.T) {
	a, b := new(LazyFelt).SetUval(^uint64(0)), new(LazyFelt).SetUval(^uint64(0))
	c := new(LazyFelt)

	c = c.Add(a, b)

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Add(new(f.Element).SetUint64(^uint64(0)), new(f.Element).SetUint64(^uint64(0)))
	assert.False(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestAddUintAndFelt(t *testing.T) {
	x := new(f.Element).SetUint64(2)
	a, b := new(LazyFelt).SetUval(0), new(LazyFelt).SetFelt(x.Inverse(x))
	c := new(LazyFelt)

	c = c.Add(a, b)

	assert.False(t, b.IsUint64())

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Add(new(f.Element).SetUint64(0), b.ToFieldElement())
	assert.False(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestAddFeltAndUint(t *testing.T) {
	x := new(f.Element).SetUint64(2)
	a, b := new(LazyFelt).SetUval(0), new(LazyFelt).SetFelt(x.Inverse(x))
	c := new(LazyFelt)

	c = c.Add(b, a)

	assert.False(t, b.IsUint64())

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Add(new(f.Element).SetUint64(0), b.ToFieldElement())
	assert.False(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestSub2Uints(t *testing.T) {
	a, b := new(LazyFelt).SetUval(50), new(LazyFelt).SetUval(20)
	c := new(LazyFelt)

	c = c.Sub(a, b)

	assert.True(t, c.IsUint64())
	assert.Equal(t, uint64(30), c.Uint64())
}

func TestSub2UintsUndeflow(t *testing.T) {
	a, b := new(LazyFelt).SetUval(0), new(LazyFelt).SetUval(20)
	c := new(LazyFelt)

	c = c.Sub(a, b)

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Sub(new(f.Element).SetUint64(0), new(f.Element).SetUint64(20))
	assert.False(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestSubUintAndFelt(t *testing.T) {
	x := new(f.Element).SetUint64(2)
	a, b := new(LazyFelt).SetUval(0), new(LazyFelt).SetFelt(x.Inverse(x))
	c := new(LazyFelt)

	c = c.Sub(a, b)

	assert.False(t, b.IsUint64())

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Sub(new(f.Element).SetUint64(0), b.ToFieldElement())
	assert.False(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestSubFeltAndUint(t *testing.T) {
	x := new(f.Element).SetUint64(2)
	a, b := new(LazyFelt).SetUval(10), new(LazyFelt).SetFelt(x.Inverse(x))
	c := new(LazyFelt)

	c = c.Sub(b, a)

	assert.False(t, b.IsUint64())

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Sub(b.ToFieldElement(), new(f.Element).SetUint64(10))
	assert.False(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestMul2UintsOverflow(t *testing.T) {
	a, b := new(LazyFelt).SetUval(^uint64(0)), new(LazyFelt).SetUval(2)
	c := new(LazyFelt)

	c = c.Mul(a, b)

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Mul(new(f.Element).SetUint64(^uint64(0)), new(f.Element).SetUint64(2))
	assert.False(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestMulUintAndFelt(t *testing.T) {
	x := new(f.Element).SetUint64(2)
	a, b := new(LazyFelt).SetUval(10), new(LazyFelt).SetFelt(x.Inverse(x))
	c := new(LazyFelt)

	c = c.Mul(a, b)

	assert.False(t, b.IsUint64())

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Mul(new(f.Element).SetUint64(10), b.ToFieldElement())
	assert.True(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}

func TestMulFeltAndUint(t *testing.T) {
	x := new(f.Element).SetUint64(2)
	a, b := new(LazyFelt).SetUval(10), new(LazyFelt).SetFelt(x.Inverse(x))
	c := new(LazyFelt)

	c = c.Mul(b, a)

	assert.False(t, b.IsUint64())

	expectedFelt := new(f.Element)
	expectedFelt = expectedFelt.Mul(b.ToFieldElement(), new(f.Element).SetUint64(10))
	assert.True(t, c.IsUint64())
	assert.Equal(t, expectedFelt, c.ToFieldElement())
}
