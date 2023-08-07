package vm

import (
	"testing"

	f "github.com/NethermindEth/juno/core/felt"
	"github.com/stretchr/testify/assert"
)

func TestAddFelt(t *testing.T) {
	r := new(Relocatable)
	r1 := NewRelocatable(2, new(f.Felt).SetUint64(10))
	r2 := new(Relocatable).SetUint64(2)
	expected := NewRelocatable(2, new(f.Felt).SetUint64(12))

	res, err := r.Add(r1, r2)

	assert.NoError(t, err)

	assert.Equal(t, r, res)
	assert.Equal(t, *res, *expected)
}

func TestAddRelocatable(t *testing.T) {
	r := new(Relocatable)
	r1 := NewRelocatable(2, new(f.Felt).SetUint64(10))
	r2 := NewRelocatable(2, new(f.Felt).SetUint64(2))

	r, err := r.Add(r1, r2)

	assert.Nil(t, r)
	assert.Error(t, err)
}

func TestSubFelt(t *testing.T) {
	r := new(Relocatable)
	r1 := NewRelocatable(2, new(f.Felt).SetUint64(10))
	r2 := new(Relocatable).SetUint64(2)
	expected := NewRelocatable(2, new(f.Felt).SetUint64(8))

	res, err := r.Sub(r1, r2)

	assert.NoError(t, err)

	assert.Equal(t, r, res)
	assert.Equal(t, *res, *expected)
}

func TestSubSameSegment(t *testing.T) {
	r := new(Relocatable)
	r1 := NewRelocatable(2, new(f.Felt).SetUint64(10))
	r2 := NewRelocatable(2, new(f.Felt).SetUint64(2))
	expected := new(Relocatable).SetUint64(8)

	res, err := r.Sub(r1, r2)

	assert.NoError(t, err)

	assert.Equal(t, r, res)
	assert.Equal(t, *res, *expected)
}

func TestSubDifferentSegment(t *testing.T) {
	r := new(Relocatable)
	r1 := NewRelocatable(2, new(f.Felt).SetUint64(10))
	r2 := NewRelocatable(1, new(f.Felt).SetUint64(2))

	r, err := r.Sub(r1, r2)

	assert.Nil(t, r)
	assert.Error(t, err)
}

func TestRelocate1(t *testing.T) {
	r := new(Relocatable)
	r1 := NewRelocatable(2, new(f.Felt).SetUint64(10))
	expected := new(Relocatable).SetUint64(52)

	res, err := r.Relocate(r1, &map[uint64]*Relocatable{
		2: new(Relocatable).SetUint64(42),
	})

	assert.NoError(t, err)
	assert.Equal(t, res, r)
	assert.Equal(t, *res, *expected)
}

func TestRelocate2(t *testing.T) {
	r := new(Relocatable)
	r1 := NewRelocatable(2, new(f.Felt).SetUint64(10))
	expected := NewRelocatable(10, new(f.Felt).SetUint64(11))

	res, err := r.Relocate(r1, &map[uint64]*Relocatable{
		2: NewRelocatable(10, new(f.Felt).SetUint64(1)),
	})

	assert.NoError(t, err)
	assert.Equal(t, res, r)
	assert.Equal(t, *res, *expected)
}

func TestRelocateMissingRule(t *testing.T) {
	r := new(Relocatable)
	r1 := NewRelocatable(2, new(f.Felt).SetUint64(10))

	res, err := r.Relocate(r1, &map[uint64]*Relocatable{
		3: NewRelocatable(10, new(f.Felt).SetUint64(1)),
	})

	assert.Error(t, err)
	assert.Nil(t, res)
}
