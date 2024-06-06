package memory

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"golang.org/x/exp/constraints"
)

// Represents a Memory Address of the Cairo VM. Because memory is split between different segments
// during execution, addresses has two locators: the segment they belong to and the location
// inside that segment
type MemoryAddress struct {
	SegmentIndex uint64
	Offset       uint64
}

var UnknownAddress = MemoryAddress{}

func (address *MemoryAddress) Equal(other *MemoryAddress) bool {
	return address.SegmentIndex == other.SegmentIndex && address.Offset == other.Offset
}

// It crates a new memory address with the modified offset
func (address *MemoryAddress) AddOffset(offset int16) (MemoryAddress, error) {
	newOffset, overflow := utils.SafeOffset(address.Offset, offset)
	if overflow {
		return UnknownAddress,
			fmt.Errorf(
				"address new invalid offseet: %d + %d = %d",
				address.Offset, offset, newOffset,
			)
	}
	return MemoryAddress{
		SegmentIndex: address.SegmentIndex,
		Offset:       newOffset,
	}, nil

}

// Adds a memory address and a field element
func (address *MemoryAddress) Add(lhs *MemoryAddress, rhs *f.Element) error {
	lhsOffset := new(f.Element).SetUint64(lhs.Offset)
	newOffset := new(f.Element).Add(lhsOffset, rhs)
	if !newOffset.IsUint64() {
		return fmt.Errorf("new offset bigger than uint64: %s", rhs.Text(10))
	}
	address.SegmentIndex = lhs.SegmentIndex
	address.Offset = newOffset.Uint64()
	return nil
}

func (address *MemoryAddress) Relocate(segmentsOffset []uint64) *f.Element {
	// no risk overflow because this sizes exists in actual Memory
	// so if by chance the uint64 addition overflowed, then we have
	// a machine with more than 2**64 bytes of memory (quite a lot!)
	return new(f.Element).SetUint64(
		segmentsOffset[address.SegmentIndex] + address.Offset,
	)
}

func (address MemoryAddress) String() string {
	return fmt.Sprintf(
		"%d:%d", address.SegmentIndex, address.Offset,
	)
}

// Stores all posible types that can be stored in a Memory cell,
//
//   - either a Felt value (an `f.Element`),
//   - or a pointer to another Memory Cell (a `MemoryAddress`)
//     both values share the same underlying memory, which is a f.Element
type MemoryValue struct {
	felt f.Element
	kind memoryValueKind
}

type memoryValueKind uint8

const (
	unknownMemoryValue memoryValueKind = iota
	feltMemoryValue
	addrMemoryValue
)

var UnknownValue = MemoryValue{}

func MemoryValueFromMemoryAddress(address *MemoryAddress) MemoryValue {
	v := MemoryValue{
		kind: addrMemoryValue,
	}
	*v.addrUnsafe() = *address
	return v
}

func MemoryValueFromFieldElement(felt *f.Element) MemoryValue {
	return MemoryValue{
		felt: *felt,
		kind: feltMemoryValue,
	}
}

func MemoryValueFromInt[T constraints.Integer](v T) MemoryValue {
	if v >= 0 {
		return MemoryValueFromUint(uint64(v))
	}

	value := MemoryValue{kind: feltMemoryValue}
	rhs := f.NewElement(uint64(-v))
	value.felt.Sub(&value.felt, &rhs)
	return value
}

func MemoryValueFromUint[T constraints.Unsigned](v T) MemoryValue {
	return MemoryValue{
		felt: f.NewElement(uint64(v)),
		kind: feltMemoryValue,
	}
}

// creates a memory value from an index and an offset. If either is negative the result is
// undefined
func MemoryValueFromSegmentAndOffset[T constraints.Integer](segmentIndex, offset T) MemoryValue {
	return MemoryValueFromMemoryAddress(
		&MemoryAddress{
			SegmentIndex: uint64(segmentIndex),
			Offset:       uint64(offset),
		},
	)
}

func MemoryValueFromAny(anyType any) (MemoryValue, error) {
	switch anyType := anyType.(type) {
	case int:
		return MemoryValueFromInt(anyType), nil
	case uint64:
		return MemoryValueFromUint(anyType), nil
	case *f.Element:
		return MemoryValueFromFieldElement(anyType), nil
	case *MemoryAddress:
		return MemoryValueFromMemoryAddress(anyType), nil
	default:
		return MemoryValue{}, fmt.Errorf("invalid type to convert to a MemoryValue: %T", anyType)
	}
}

func EmptyMemoryValueAsFelt() MemoryValue {
	return MemoryValue{
		kind: feltMemoryValue,
	}
}

func EmptyMemoryValueAsAddress() MemoryValue {
	return MemoryValue{
		kind: addrMemoryValue,
	}
}

func EmptyMemoryValueAs(address bool) MemoryValue {
	kind := feltMemoryValue
	if address {
		kind = addrMemoryValue
	}
	return MemoryValue{
		kind: kind,
	}
}

func (mv *MemoryValue) MemoryAddress() (*MemoryAddress, error) {
	if !mv.IsAddress() {
		return nil, errors.New("memory value is not an address")
	}
	return mv.addrUnsafe(), nil
}

func (mv *MemoryValue) FieldElement() (*f.Element, error) {
	if !mv.IsFelt() {
		return nil, fmt.Errorf("memory value is not a field element")
	}
	return &mv.felt, nil
}

func (mv *MemoryValue) Any() any {
	if mv.IsAddress() {
		return mv.addrUnsafe()
	}
	return &mv.felt
}

func (mv *MemoryValue) IsAddress() bool {
	return mv.kind == addrMemoryValue
}

func (mv *MemoryValue) IsFelt() bool {
	return mv.kind == feltMemoryValue
}

func (mv *MemoryValue) Known() bool {
	return mv.kind != unknownMemoryValue
}

func (mv *MemoryValue) Equal(other *MemoryValue) bool {
	if mv.IsAddress() && other.IsAddress() {
		return mv.addrUnsafe().Equal(other.addrUnsafe())
	}
	if mv.IsFelt() && other.IsFelt() {
		return mv.felt.Equal(&other.felt)
	}
	return false
}

// Adds two memory values is the second one is a Felt
func (mv *MemoryValue) Add(lhs, rhs *MemoryValue) error {
	if lhs.IsAddress() {
		if !rhs.IsFelt() {
			return errors.New("rhs is not a felt")
		}
		return mv.addrUnsafe().Add(lhs.addrUnsafe(), &rhs.felt)
	}
	if rhs.IsAddress() {
		return mv.addrUnsafe().Add(rhs.addrUnsafe(), &lhs.felt)
	}

	mv.felt.Add(&lhs.felt, &rhs.felt)
	return nil
}

// Subs two memory values if they're in the same segment or the rhs is a Felt.
func (mv *MemoryValue) Sub(lhs, rhs *MemoryValue) error {
	if lhs.IsAddress() {
		return mv.subAddress(lhs.addrUnsafe(), rhs)
	}

	if rhs.IsAddress() {
		return errors.New("cannot substract an address from a felt")
	}

	mv.felt.Sub(&lhs.felt, &rhs.felt)
	return nil
}

// subAddress subtracts from a memory address a felt or another memory address in the same segment.
func (mv *MemoryValue) subAddress(lhs *MemoryAddress, rhs *MemoryValue) error {
	// There are only two supported forms of this operation:
	// * addr sub addr => offset as felt
	// * adds sub felt => addr

	if rhs.IsAddress() {
		// The result is the offset value, felt-typed.
		// Both addresses need to belong to the same segment.
		// See #284
		rhsAddr := rhs.addrUnsafe()
		if lhs.SegmentIndex != rhsAddr.SegmentIndex {
			return fmt.Errorf("addresses are in different segments: rhs is in %d, lhs is in %d",
				rhsAddr.SegmentIndex, lhs.SegmentIndex)
		}
		mv.kind = feltMemoryValue
		mv.felt.SetUint64(lhs.Offset - rhsAddr.Offset)
		return nil
	}

	// rhs is felt, the result is address.
	if !rhs.felt.IsUint64() {
		return fmt.Errorf("rhs field element does not fit in uint64: %s", &rhs.felt)
	}
	rhs64 := rhs.felt.Uint64()
	if rhs64 > lhs.Offset {
		return fmt.Errorf("rhs %d is greater than lhs offset %d", rhs64, lhs.Offset)
	}
	mv.kind = addrMemoryValue
	addrResult := mv.addrUnsafe()
	addrResult.SegmentIndex = lhs.SegmentIndex
	addrResult.Offset = lhs.Offset - rhs64
	return nil
}

func (mv *MemoryValue) Mul(lhs, rhs *MemoryValue) error {
	if lhs.IsAddress() || rhs.IsAddress() {
		return errors.New("cannot multiply memory addresses")
	}
	mv.felt.Mul(&lhs.felt, &rhs.felt)
	return nil
}

func (mv *MemoryValue) Div(lhs, rhs *MemoryValue) error {
	if lhs.IsAddress() || rhs.IsAddress() {
		return errors.New("cannot divide memory addresses")
	}
	mv.felt.Div(&lhs.felt, &rhs.felt)
	return nil
}

func (mv MemoryValue) String() string {
	if mv.IsAddress() {
		return mv.addrUnsafe().String()
	}
	return mv.felt.String()
}

// Retuns a MemoryValue holding a felt as uint if it fits
func (mv *MemoryValue) Uint64() (uint64, error) {
	if mv.IsAddress() {
		return 0, fmt.Errorf("cannot convert a memory address into uint64: %s", *mv)
	}
	if !mv.felt.IsUint64() {
		return 0, fmt.Errorf("field element does not fit in uint64: %s", mv.String())
	}

	return mv.felt.Uint64(), nil
}

func (mv *MemoryValue) addrUnsafe() *MemoryAddress {
	return (*MemoryAddress)(unsafe.Pointer(&mv.felt))
}

func (memory *Memory) GetConsecutiveMemoryValues(addr MemoryAddress, size int16) ([]MemoryValue, error) {
	values := make([]MemoryValue, size)
	for i := int16(0); i < size; i++ {
		nAddr, err := addr.AddOffset(i)
		if err != nil {
			return nil, err
		}
		v, err := memory.ReadFromAddress(&nAddr)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return values, nil
}

func (memory *Memory) ResolveAsBigInt3(valAddr MemoryAddress) ([3]*f.Element, error) {
	valMemoryValues, err := memory.GetConsecutiveMemoryValues(valAddr, int16(3))
	if err != nil {
		return [3]*f.Element{}, err
	}

	var valValues [3]*f.Element
	for i := 0; i < 3; i++ {
		valValue, err := valMemoryValues[i].FieldElement()
		if err != nil {
			return [3]*f.Element{}, err
		}
		valValues[i] = valValue
	}

	return valValues, nil
}
