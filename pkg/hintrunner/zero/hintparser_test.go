package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/stretchr/testify/require"
)

func TestHintParser(t *testing.T) {
	type testSetType struct {
		Parameter         string
		ExpectedReference hinter.Reference
	}

	testSet := []testSetType{
		{
			Parameter:         "cast(fp + (-3), felt*)",
			ExpectedReference: hinter.FpCellRef(-3),
		},
		{
			Parameter: "[cast(ap + (-1) + 2, starkware.cairo.common.cairo_builtins.BitwiseBuiltin**)]",
			ExpectedReference: hinter.Deref{
				Deref: hinter.ApCellRef(1),
			},
		},
		{
			Parameter: "[cast([ap + 2], felt)]",
			ExpectedReference: hinter.DoubleDeref{
				Deref: hinter.Deref{
					Deref: hinter.ApCellRef(2),
				},
				Offset: 0,
			},
		},
		{
			Parameter: "cast([ap + 2] + [ap], felt)",
			ExpectedReference: hinter.BinaryOp{
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
			Parameter: "cast([ap + (-5)] * [ap + (-1)], felt)",
			ExpectedReference: hinter.BinaryOp{
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
			Parameter: "cast([ap] * 3, felt)",
			ExpectedReference: hinter.BinaryOp{
				Operator: hinter.Mul,
				Lhs: hinter.Deref{
					Deref: hinter.ApCellRef(0),
				},
				Rhs: hinter.Immediate{18446744073709551521, 18446744073709551615, 18446744073709551615, 576460752303421872},
			},
		},
		{
			Parameter:         "cast(2389472938759290879897, felt)",
			ExpectedReference: hinter.Immediate(*feltString("2389472938759290879897")),
		},
		{
			Parameter: "cast([[ap + 2] + (-5)], felt)",
			ExpectedReference: hinter.DoubleDeref{
				Deref: hinter.Deref{
					Deref: hinter.ApCellRef(2),
				},
				Offset: int16(-5),
			},
		},
		{
			Parameter: "cast([fp + (-4)] * 18, felt)",
			ExpectedReference: hinter.BinaryOp{
				Operator: hinter.Mul,
				Lhs: hinter.Deref{
					Deref: hinter.FpCellRef(-4),
				},
				Rhs: hinter.Immediate(*feltInt64(18)),
			},
		},
		{
			Parameter: "[cast(ap - 0 + (-1), felt*)]",
			ExpectedReference: hinter.Deref{
				Deref: hinter.ApCellRef(-1),
			},
		},
	}

	for _, test := range testSet {
		output, err := ParseIdentifier(test.Parameter)
		require.NoError(t, err)

		if test.ExpectedReference != nil {
			require.Equal(t, test.ExpectedReference, output, "unexpected Reference type")
		}
	}
}
