package hintrunner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHintParser(t *testing.T) {
	type testSetType struct {
		Parameter            string
		ExpectedCellRefer    CellRefer
		ExpectedResOperander ResOperander
	}

	testSet := []testSetType{
		{
			Parameter:            "cast(fp + (-3), felt*)",
			ExpectedCellRefer:    FpCellRef(-3),
			ExpectedResOperander: nil,
		},
		{
			Parameter:         "[cast(ap + (-1) + 2, starkware.cairo.common.cairo_builtins.BitwiseBuiltin**)]",
			ExpectedCellRefer: nil,
			ExpectedResOperander: Deref{
				deref: ApCellRef(1),
			},
		},
		{
			Parameter:         "[cast([ap + 2], felt)]",
			ExpectedCellRefer: nil,
			ExpectedResOperander: DoubleDeref{
				deref:  ApCellRef(2),
				offset: 0,
			},
		},
		{
			Parameter:         "cast([ap + 2] + [ap], felt)",
			ExpectedCellRefer: nil,
			ExpectedResOperander: BinaryOp{
				operator: Add,
				lhs:      ApCellRef(2),
				rhs: Deref{
					deref: ApCellRef(0),
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
