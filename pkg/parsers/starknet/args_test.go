package starknet

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStarknetProgramArgs(t *testing.T) {
	testCases := []struct {
		name     string
		args     string
		expected []CairoFuncArgs
		err      error
	}{
		{
			name: "single arg",
			args: "1",
			expected: []CairoFuncArgs{
				{
					Single: new(fp.Element).SetUint64(1),
					Array:  nil,
				},
			},
			err: nil,
		},
		{
			name: "single array arg",
			args: "[1 2 3 4]",
			expected: []CairoFuncArgs{
				{
					Single: nil,
					Array: []fp.Element{
						*new(fp.Element).SetUint64(1),
						*new(fp.Element).SetUint64(2),
						*new(fp.Element).SetUint64(3),
						*new(fp.Element).SetUint64(4),
					},
				},
			},
		},
		{
			name:     "nested array arg",
			args:     "[1 [2 3 4]]",
			expected: nil,
			err:      fmt.Errorf("invalid felt value in array: invalid felt value: Element.SetString failed -> can't parse number into a big.Int [2"),
		},
		{
			name: "mixed args",
			args: "1 [2 3 4] 5 [6 7 8] [1] 9 9 [12341341234 0]",
			expected: []CairoFuncArgs{
				{
					Single: new(fp.Element).SetUint64(1),
					Array:  nil,
				},
				{
					Single: nil,
					Array: []fp.Element{
						*new(fp.Element).SetUint64(2),
						*new(fp.Element).SetUint64(3),
						*new(fp.Element).SetUint64(4),
					},
				},
				{
					Single: new(fp.Element).SetUint64(5),
					Array:  nil,
				},
				{
					Single: nil,
					Array: []fp.Element{
						*new(fp.Element).SetUint64(6),
						*new(fp.Element).SetUint64(7),
						*new(fp.Element).SetUint64(8),
					},
				},
				{
					Single: nil,
					Array: []fp.Element{
						*new(fp.Element).SetUint64(1),
					},
				},
				{
					Single: new(fp.Element).SetUint64(9),
					Array:  nil,
				},
				{
					Single: new(fp.Element).SetUint64(9),
					Array:  nil,
				},
				{
					Single: nil,
					Array: []fp.Element{
						*new(fp.Element).SetUint64(12341341234),
						*new(fp.Element).SetUint64(0),
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			args, err := ParseCairoProgramArgs(testCase.args)
			if testCase.err != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.err.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, args)
		})
	}
}
