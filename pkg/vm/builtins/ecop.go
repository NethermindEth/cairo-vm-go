package builtins

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

// todo:
// * Set alpha and beta felt values
// * Set feltTwo and feltThree value

const EcOpName = "ec_op"
const cellsPerEcOp = 7
const inputCellsPerEcOp = 5

type EcOp struct{}

func (e *EcOp) CheckWrite(segment *mem.Segment, offset uint64, value *mem.MemoryValue) error {
	return nil
}

func (e *EcOp) InferValue(segment *mem.Segment, offset uint64) error {
	// get the current slot index and verify it is an output cell
	ecopIndex := offset % cellsPerEcOp
	if ecopIndex < inputCellsPerEcOp {
		return errors.New("cannot infer value from input cell")
	}

	// gather the input cells
	inputOff := offset - ecopIndex
	inputs := [5]mem.MemoryValue{
		segment.Peek(inputOff),
		segment.Peek(inputOff + 1),
		segment.Peek(inputOff + 2),
		segment.Peek(inputOff + 3),
		segment.Peek(inputOff + 4),
	}

	// assert all values are known
	for i := range inputs {
		if !inputs[i].Known() {
			return fmt.Errorf(
				"cannot infer value: input value at offset %d is unknown", inputOff+uint64(i),
			)
		}
	}

	// unwrap the values as felts
	inputsFelt := [5]*f.Element{}
	for i := range inputs {
		felt, err := inputs[i].FieldElement()
		if err != nil {
			return err
		}
		inputsFelt[i] = felt
	}

	// todo(rodro): the python vm has an upper limit on the size of `m`(the fifth input)  but
	// since the max size seems to be maxout at 2**252, I see no point to currently add a check
	// for it

	// verify p and q are in the curve
	var alpha, beta f.Element
	p := point{*inputsFelt[0], *inputsFelt[1]}
	q := point{*inputsFelt[2], *inputsFelt[3]}
	if !pointOnCurve(&p, &alpha, &beta) {
		return fmt.Errorf("point P(%s, %s) is not on the curve", &p.X, &p.Y)
	}
	if !pointOnCurve(&q, &alpha, &beta) {
		return fmt.Errorf("point Q(%s, %s) is not on the curve", &q.X, &q.Y)
	}

	r, err := ecop(&p, &q, inputsFelt[4], &alpha, &beta)
	if err != nil {
		return err
	}

	outputOff := inputOff + inputCellsPerEcOp

	rxMV := mem.MemoryValueFromFieldElement(&r.X)
	err = segment.Write(outputOff, &rxMV)
	if err != nil {
		return err
	}

	ryMV := mem.MemoryValueFromFieldElement(&r.Y)
	err = segment.Write(outputOff+1, &ryMV)
	return err
}

// structure to represent a point in the elliptic curve
type point struct {
	X, Y f.Element
}

// returns true if a point `p` ins the `ec` curve ruled by the params `alpha`,
// `beta` and `prime`. In other words, true if  y^2 = x^3 + alpha * x + beta
func pointOnCurve(p *point, alpha, beta *f.Element) bool {
	// calculate lhs
	y2 := f.Element{}
	y2.Square(&p.Y)

	// calculate rhs
	x3 := f.Element{}
	x3.Square(&p.X)
	x3.Mul(&x3, &p.X)

	ax := f.Element{}
	ax.Mul(alpha, &p.X)

	x3.Add(&x3, &ax)
	x3.Add(&x3, beta)

	// return lhs == rhs
	return y2.Equal(&x3)
}

// returns the result of the ecop operation on points `P` and `Q` with params
// `m`, `alpha` and `beta`. The resulting point `R` is equal to  P + m * Q
func ecop(p *point, q *point, m, alpha, beta *f.Element) (point, error) {
	doublePoint := *p
	partialSum := *q

	mBytes := m.Bytes()
	scalar := uint256.Int{}
	scalar.SetBytes32(mBytes[:])

	// Note: In the python VM the height is a parameter but it is always set at 256,
	// we treated as a constant as a result
	const height = 256
	for i := 0; i < height; i++ {
		// we check that both points are always different between each others
		// `ecadd` assume `x` ordinates are always different
		// `ecdouble` assumes `y` coordinates are always different
		if doublePoint.X.Equal(&partialSum.X) || doublePoint.Y.Equal(&partialSum.Y) {
			return point{}, fmt.Errorf(
				"at least one similar coordinates between P(%s, %s) and Q(%s, %s)",
				&p.X, &p.Y, &q.X, &q.Y,
			)
		}
		and := uint256.Int{}
		and.And(&scalar, &utils.Uint256One)
		if and.Cmp(&utils.Uint256Zero) != 0 {
			partialSum = ecadd(&partialSum, &doublePoint)
		}

		doublePoint = ecdouble(&doublePoint, alpha)
		scalar.Rsh(&scalar, 1)
	}

	return partialSum, nil
}

// performs elliptic curve addition over two points. Assumes `x` ordinates are
// always different
func ecadd(p *point, q *point) point {
	// get the slope between the two points
	slope := f.Element{}
	slope.Sub(&p.Y, &q.Y)
	denom := f.Element{}
	denom.Sub(&p.X, &q.X)
	slope.Div(&slope, &denom)

	// get the x coordinate: x = slope^2 - p.X - q.X
	x := f.Element{}
	x.Square(&slope)
	x.Sub(&x, &p.X)
	x.Sub(&x, &q.X)

	// get the y coordinate: y = slope * (p.X - x) - p.Y
	y := f.Element{}
	y.Sub(&p.X, &x)
	y.Mul(&y, &slope)
	y.Sub(&y, &p.Y)

	return point{x, y}
}

// performs elliptic curve doubling over a point. Assumes `y` ordinates are
// always different
func ecdouble(p *point, alpha *f.Element) point {
	// get the double slope
	doubleSlope := f.Element{}
	doubleSlope.Square(&p.X)
	doubleSlope.Mul(&doubleSlope, &utils.FeltThree)
	doubleSlope.Add(&doubleSlope, alpha)

	// get the x coordinate: x = slope^2 - 2 * p.X
	x := f.Element{}
	x.Square(&doubleSlope)
	doublePx := f.Element{}
	doublePx.Double(&p.X)
	x.Sub(&x, &doublePx)

	// get the y coordinates: y =  slope * (p.X - x) - p.Y
	y := f.Element{}
	y.Sub(&p.X, &x)
	y.Mul(&y, &doubleSlope)
	y.Sub(&y, &p.Y)

	return point{x, y}
}
