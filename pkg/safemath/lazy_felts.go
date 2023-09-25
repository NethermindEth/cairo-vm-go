package safemath

import (
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const (
	lazyUint        = 0x1
	lazyFelt        = 0x2
	lazyUintAndFelt = lazyUint | lazyFelt
)

type LazyFelt struct {
	mask    uint8
	uval    uint64
	feltval *f.Element
}

func (x *LazyFelt) SetFelt(felt *f.Element) *LazyFelt {
	x.mask = lazyFelt
	x.feltval = felt

	return x
}

func (x *LazyFelt) SetUval(val uint64) *LazyFelt {
	x.mask = lazyUint
	x.uval = val

	return x
}

func (x *LazyFelt) allocFelt() *LazyFelt {
	if x.feltval == nil {
		x.mask = lazyFelt
		x.feltval = new(f.Element)
	}

	return x
}

func (z *LazyFelt) Add(x, y *LazyFelt) *LazyFelt {

	if x.mask&y.mask&lazyUint != 0 {
		res, isOverflow := AddAndCheckOverflow(x.uval, y.uval)
		if isOverflow {
			xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
			z.allocFelt().feltval.Add(xFelt, yFelt)
		} else {
			z.SetUval(res)
		}
	} else {
		xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
		z.allocFelt().feltval.Add(xFelt, yFelt)
	}

	return z
}

func (z *LazyFelt) Sub(x, y *LazyFelt) *LazyFelt {

	if x.mask&y.mask&lazyUint != 0 {
		res, isOverflow := SubAndCheckUnderflow(x.uval, y.uval)
		if isOverflow {
			xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
			z.allocFelt().feltval.Sub(xFelt, yFelt)
		} else {
			z.SetUval(res)
		}
	} else {
		xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
		z.allocFelt().feltval.Sub(xFelt, yFelt)
	}

	return z
}

func (z *LazyFelt) Mul(x, y *LazyFelt) *LazyFelt {

	if x.mask&y.mask&lazyUint != 0 {
		res, isOverflow := MulAndCheckOverflow(x.uval, y.uval)
		if isOverflow {
			xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
			z.allocFelt().feltval.Mul(xFelt, yFelt)
		} else {
			z.SetUval(res)
		}
	} else {
		xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
		z.allocFelt().feltval.Mul(xFelt, yFelt)
	}

	return z
}

func (z *LazyFelt) Div(x, y *LazyFelt) *LazyFelt {

	if x.mask&y.mask&lazyUint != 0 {
		if x.uval%y.uval == 0 {
			z.SetUval(x.uval / y.uval)
		} else {
			xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
			z.allocFelt().feltval.Div(xFelt, yFelt)
		}
	} else {
		xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
		z.allocFelt().feltval.Div(xFelt, yFelt)
	}

	return z
}

func (x *LazyFelt) Equal(y *LazyFelt) bool {
	if x.mask&y.mask&lazyUint != 0 {
		return x.uval == y.uval
	} else {
		xFelt, yFelt := x.ToFieldElement(), y.ToFieldElement()
		return xFelt.Equal(yFelt)
	}
}

func (x *LazyFelt) IsZero() bool {
	if x.mask&lazyUint != 0 {
		return x.uval == 0
	} else {
		return x.feltval.IsZero()
	}
}

func (x *LazyFelt) IsUint64() bool {
	if x.mask&lazyUint != 0 {
		return true
	} else {
		return x.feltval.IsUint64()
	}
}

func (x *LazyFelt) Uint64() uint64 {
	if x.mask&lazyUint != 0 {
		return x.uval
	} else {
		x.uval = x.feltval.Uint64()
		x.mask |= lazyUint
		return x.uval
	}
}

func (x *LazyFelt) String() string {
	if x.mask&lazyUint != 0 {
		return fmt.Sprintf("%d", x.uval)
	} else {
		return x.feltval.String()
	}
}

func (x *LazyFelt) ToFieldElement() *f.Element {
	if x.mask&lazyFelt == 0 {
		x.mask |= lazyFelt
		x.feltval = new(f.Element).SetUint64(x.uval)
	}

	return x.feltval
}
