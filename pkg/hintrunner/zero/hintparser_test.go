package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/stretchr/testify/require"
)

func TestHintParser(t *testing.T) {
	type testSetType struct {
		Parameter            string
		ExpectedCellRefer    hinter.CellRefer
		ExpectedResOperander hinter.ResOperander
	}

	testSet := []testSetType{
		{
			Parameter:            "cast(fp + (-3), felt*)",
			ExpectedCellRefer:    hinter.FpCellRef(-3),
			ExpectedResOperander: nil,
		},
		{
			Parameter:         "[cast(ap + (-1) + 2, starkware.cairo.common.cairo_builtins.BitwiseBuiltin**)]",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.Deref{
				Deref: hinter.ApCellRef(1),
			},
		},
		{
			Parameter:         "[cast([ap + 2], felt)]",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.DoubleDeref{
				Deref: hinter.Deref{
					Deref: hinter.ApCellRef(2),
				},
				Offset: 0,
			},
		},
		{
			Parameter:         "cast([ap + 2] + [ap], felt)",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.BinaryOp{
				Operator: hinter.Add,
				Lhs: hinter.Deref{
					Deref: hinter.ApCellRef(2),
				},
				Rhs: hinter.Deref{
					Deref: hinter.ApCellRef(0),
				},
			},
		},
		{
			Parameter:         "cast([ap + (-5)] * [ap + (-1)], felt)",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.BinaryOp{
				Operator: hinter.Mul,
				Lhs: hinter.Deref{
					Deref: hinter.ApCellRef(-5),
				},
				Rhs: hinter.Deref{
					Deref: hinter.ApCellRef(-1),
				},
			},
		},
		{
			Parameter:         "cast([ap] * 3, felt)",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.BinaryOp{
				Operator: hinter.Mul,
				Lhs: hinter.Deref{
					Deref: hinter.ApCellRef(0),
				},
				Rhs: hinter.Immediate{18446744073709551521, 18446744073709551615, 18446744073709551615, 576460752303421872},
			},
		},
		{
			Parameter:            "cast(2389472938759290879897, felt)",
			ExpectedCellRefer:    nil,
			ExpectedResOperander: hinter.Immediate(*feltString("2389472938759290879897")),
		},
		{
			Parameter:         "cast([[ap + 2] + (-5)], felt)",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.DoubleDeref{
				Deref: hinter.Deref{
					Deref: hinter.ApCellRef(2),
				},
				Offset: int16(-5),
			},
		},
		{
			Parameter:         "cast([fp + (-4)] * 18, felt)",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.BinaryOp{
				Operator: hinter.Mul,
				Lhs: hinter.Deref{
					Deref: hinter.FpCellRef(-4),
				},
				Rhs: hinter.Immediate(*feltInt64(18)),
			},
		},
		{
			Parameter:         "[cast(ap - 0 + (-1), felt*)]",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.Deref{
				Deref: hinter.ApCellRef(-1),
			},
		},
	}

	for _, test := range testSet {
		output, err := ParseIdentifier(test.Parameter)
		require.NoError(t, err)

		if test.ExpectedCellRefer != nil {
			require.Equal(t, test.ExpectedCellRefer, output, "unexpected CellRefer type")
		}

		if test.ExpectedResOperander != nil {
			require.Equal(t, test.ExpectedResOperander, output, "unexpected ResOperander type")
		}
	}
}
