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
				Deref:  hinter.ApCellRef(2),
				Offset: 0},
		},
		{
			Parameter:         "cast([ap + 2] + [ap], felt)",
			ExpectedCellRefer: nil,
			ExpectedResOperander: hinter.BinaryOp{
				Operator: hinter.Add,
				Lhs:      hinter.ApCellRef(2),
				Rhs: hinter.Deref{
					Deref: hinter.ApCellRef(0),
				},
			},
		},
	}

	for _, test := range testSet {
		output, err := ParseIdentifier(test.Parameter)
		require.NoError(t, err)

		if test.ExpectedCellRefer != nil {
			require.Equal(t, test.ExpectedCellRefer, output, "Expected CellRefer type")
		}

		if test.ExpectedResOperander != nil {
			require.Equal(t, test.ExpectedResOperander, output, "Expected ResOperander type")
		}
	}
}
