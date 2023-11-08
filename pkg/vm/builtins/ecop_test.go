package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestEcOp(t *testing.T) {
	// input p
	px, _ := new(fp.Element).SetString("0x49EE3EBA8C1600700EE1B87EB599F16716B0B1022947733551FDE4050CA6804")
	py, _ := new(fp.Element).SetString("0x3CA0CFE4B3BC6DDF346D49D06EA0ED34E621062C0E056C1D0405D266E10268A")

	// input q
	qx, _ := new(fp.Element).SetString("0x1EF15C18599971B7BECED415A40F0C7DEACFD9B0D1819E03D723D8BC943CFCA")

	qy, _ := new(fp.Element).SetString("0x5668060AA49730B7BE4801DF46EC62DE53ECD11ABE43A32873000C36E8DC1F")

	// input m
	m := new(fp.Element).SetInt64(3)

	// expected r
	mult := ecmult(&point{*qx, *qy}, new(uint256.Int).SetUint64(3), &utils.FeltOne)
	r := ecadd(&point{*px, *py}, &mult)

	segment := memory.EmptySegmentWithLength(cellsPerEcOp)
	ecop := &EcOp{}
	segment.WithBuiltinRunner(ecop)

	// write P to segment
	pxValue := memory.MemoryValueFromFieldElement(px)
	require.NoError(t, segment.Write(0, &pxValue))
	pyValue := memory.MemoryValueFromFieldElement(py)
	require.NoError(t, segment.Write(1, &pyValue))

	// write Q to segment
	qxValue := memory.MemoryValueFromFieldElement(qx)
	require.NoError(t, segment.Write(2, &qxValue))
	qyValue := memory.MemoryValueFromFieldElement(qy)
	require.NoError(t, segment.Write(3, &qyValue))

	// write m to segment
	mValue := memory.MemoryValueFromFieldElement(m)
	require.NoError(t, segment.Write(4, &mValue))

	// infer R
	rxValue, err := segment.Read(5)
	require.NoError(t, err)
	ryValue, err := segment.Read(6)
	require.NoError(t, err)

	rx, err := rxValue.FieldElement()
	require.NoError(t, err)
	ry, err := ryValue.FieldElement()
	require.NoError(t, err)

	require.Equal(t, r.X, *rx)
	require.Equal(t, r.Y, *ry)
}

// performs elliptic curve multiplication on point `p` with scalar `m` and param `alpha`.
// `m` value gets modified in place
func ecmult(p *point, m *uint256.Int, alpha *fp.Element) point {
	if m.Eq(&utils.Uint256One) {
		return *p
	}

	// if m is even
	and := uint256.Int{}
	and.And(m, &utils.Uint256One)
	if and.IsZero() {
		m.Rsh(m, 1)
		double := ecdouble(p, alpha)
		return ecmult(&double, m, alpha)
	}

	m.SubUint64(m, 1)
	mult := ecmult(p, m, alpha)
	return ecadd(p, &mult)
}
