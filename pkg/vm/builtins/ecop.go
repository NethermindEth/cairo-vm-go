package builtins

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
)

const EcOpName = "ec_op"
const cellsPerEcOp = 7
const inputCellsPerEcOp = 5
const instancesPerComponentEcOp = 1

var feltThree fp.Element = fp.Element(
	[]uint64{
		18446744073709551521,
		18446744073709551615,
		18446744073709551615,
		576460752303421872,
	})

type EcOp struct {
	ratio uint64
	cache map[uint64]fp.Element
}

func (e *EcOp) String() string {
	return EcOpName
}

func (e *EcOp) CheckWrite(segment *mem.Segment, offset uint64, value *mem.MemoryValue) error {
	return nil
}

func (e *EcOp) InferValue(segment *mem.Segment, offset uint64) error {
	// check if the value is already in the cache
	value, ok := e.cache[offset]
	if ok {
		mv := mem.MemoryValueFromFieldElement(&value)
		return segment.Write(offset, &mv)
	}
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
	inputsFelt := [5]*fp.Element{}
	for i := range inputs {
		felt, err := inputs[i].FieldElement()
		if err != nil {
			return err
		}
		inputsFelt[i] = felt
	}

	// Note: the python vm has an upper limit on the size of `m`(the fifth input)  but
	// since it is always maxout at 2**252, I see no point on adding a check
	// for it for now

	// verify p and q are in the curve
	p := point{*inputsFelt[0], *inputsFelt[1]}
	q := point{*inputsFelt[2], *inputsFelt[3]}
	if !p.onCurve(&utils.Alpha, &utils.Beta) {
		return fmt.Errorf("point P(%s, %s) is not on the curve", &p.X, &p.Y)
	}
	if !q.onCurve(&utils.Alpha, &utils.Beta) {
		return fmt.Errorf("point Q(%s, %s) is not on the curve", &q.X, &q.Y)
	}

	// calculate the elliptic curve operation
	r, err := ecop(&p, &q, inputsFelt[4], &utils.Alpha)
	if err != nil {
		return err
	}

	// store the resulting point `r`
	outputOff := inputOff + inputCellsPerEcOp

	// store the x and y coordinates of the resulting point
	e.cache[outputOff] = r.X
	e.cache[outputOff+1] = r.Y

	value = e.cache[offset]
	mv := mem.MemoryValueFromFieldElement(&value)
	return segment.Write(offset, &mv)
}

func (e *EcOp) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, e.ratio, inputCellsPerEcOp, instancesPerComponentEcOp, cellsPerEcOp)
}

// structure to represent a point in the elliptic curve
type point struct {
	X, Y fp.Element
}

// returns true if a point `p` belongs to the `ec` curve ruled by the params `alpha` and
// `beta`. In other words, true if  y^2 = x^3 + alpha * x + beta
func (p *point) onCurve(alpha, beta *fp.Element) bool {
	// calculate lhs
	y2 := fp.Element{}
	y2.Square(&p.Y)

	// calculate rhs
	x3 := fp.Element{}
	x3.Square(&p.X)
	x3.Mul(&x3, &p.X)

	ax := fp.Element{}
	ax.Mul(alpha, &p.X)

	x3.Add(&x3, &ax)
	x3.Add(&x3, beta)

	// return lhs == rhs
	return y2.Equal(&x3)
}

// returns the result of the ecop operation on points `P` and `Q` with scalar
// `m` and param `alpha`. The resulting point `R` is equal to  P + m * Q
func ecop(p *point, q *point, m, alpha *fp.Element) (point, error) {
	partialSum := *p
	doublePoint := *q

	mBytes := m.Bytes()
	scalar := uint256.Int{}
	scalar.SetBytes32(mBytes[:])

	// Note: In the python VM the height is a parameter but it is always set at 256
	// therefore we treat it as a constant
	const height = 256
	// todo(rodro): iteration could be cut short on the biggest bit with a one of the `scalar`
	for i := 0; i < height && !scalar.IsZero(); i++ {
		// we check that both points are always different between each others
		// `ecadd` assume `x` ordinates are always different
		// `ecdouble` assumes `y` coordinates are always different
		if doublePoint.X.Equal(&partialSum.X) || doublePoint.Y.Equal(&utils.FeltZero) {
			return point{}, fmt.Errorf(
				"EcOp requires from P(%s, %s) and Q(%s, %s) that P.X != Q.X and Q.Y != 0 ",
				&p.X, &p.Y, &q.X, &q.Y,
			)
		}
		and := uint256.Int{}
		and.And(&scalar, &utils.Uint256One)
		if !and.IsZero() {
			partialSum = ecadd(&partialSum, &doublePoint)
		}

		// todo(rodro): This loop can be optimized, potentially innecesary shift operations
		doublePoint = ecdouble(&doublePoint, alpha)
		scalar.Rsh(&scalar, 1)
	}

	return partialSum, nil
}

// performs elliptic curve addition over two points. Assumes `x` ordinates are
// always different
func ecadd(p *point, q *point) point {
	// get the slope between the two points
	slope := fp.Element{}
	slope.Sub(&p.Y, &q.Y)
	denom := fp.Element{}
	denom.Sub(&p.X, &q.X)
	slope.Div(&slope, &denom)

	// get the x coordinate: x = slope^2 - p.X - q.X
	x := fp.Element{}
	x.Square(&slope)
	x.Sub(&x, &p.X)
	x.Sub(&x, &q.X)

	// get the y coordinate: y = slope * (p.X - x) - p.Y
	y := fp.Element{}
	y.Sub(&p.X, &x)
	y.Mul(&y, &slope)
	y.Sub(&y, &p.Y)

	return point{x, y}
}

// performs elliptic curve doubling over a point. Assumes `y` coordinate
// is different than 0
func ecdouble(p *point, alpha *fp.Element) point {
	// get the double slope
	doubleSlope := fp.Element{}
	doubleSlope.Square(&p.X)
	doubleSlope.Mul(
		&doubleSlope,
		&feltThree,
	)
	doubleSlope.Add(&doubleSlope, alpha)
	denom := fp.Element{}
	denom.Double(&p.Y)
	doubleSlope.Div(&doubleSlope, &denom)

	// get the x coordinate: x = slope^2 - 2 * p.X
	x := fp.Element{}
	x.Square(&doubleSlope)
	doublePx := fp.Element{}
	doublePx.Double(&p.X)
	x.Sub(&x, &doublePx)

	// get the y coordinates: y =  slope * (p.X - x) - p.Y
	y := fp.Element{}
	y.Sub(&p.X, &x)
	y.Mul(&y, &doubleSlope)
	y.Sub(&y, &p.Y)

	return point{x, y}
}

type AirPrivateBuiltinEcOp struct {
	Index int    `json:"index"`
	PX    string `json:"p_x"`
	PY    string `json:"p_y"`
	M     string `json:"m"`
	QX    string `json:"q_x"`
	QY    string `json:"q_y"`
}

func (e *EcOp) GetAirPrivateInput(ecOpSegment *mem.Segment) []AirPrivateBuiltinEcOp {
	valueMapping := make(map[int]AirPrivateBuiltinEcOp)
	for index, value := range ecOpSegment.Data {
		if !value.Known() {
			continue
		}
		idx, typ := index/cellsPerEcOp, index%cellsPerEcOp
		if typ >= inputCellsPerEcOp {
			continue
		}

		builtinValue, exists := valueMapping[idx]
		if !exists {
			builtinValue = AirPrivateBuiltinEcOp{Index: idx}
		}

		valueBig := big.Int{}
		value.Felt.BigInt(&valueBig)
		valueHex := fmt.Sprintf("0x%x", &valueBig)
		if typ == 0 {
			builtinValue.PX = valueHex
		} else if typ == 1 {
			builtinValue.PY = valueHex
		} else if typ == 2 {
			builtinValue.QX = valueHex
		} else if typ == 3 {
			builtinValue.QY = valueHex
		} else if typ == 4 {
			builtinValue.M = valueHex
		}
		valueMapping[idx] = builtinValue
	}

	values := make([]AirPrivateBuiltinEcOp, 0)

	sortedIndexes := make([]int, 0, len(valueMapping))
	for index := range valueMapping {
		sortedIndexes = append(sortedIndexes, index)
	}
	sort.Ints(sortedIndexes)
	for _, index := range sortedIndexes {
		value := valueMapping[index]
		values = append(values, value)
	}
	return values
}
