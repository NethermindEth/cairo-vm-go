package zero

import (
	"testing"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestLoadCairoZeroProgram(t *testing.T) {
	content := []byte(`
        {
            "data": [
                "0x0000001",
                "0x0000002",
                "0x0000003",
                "0x0000004"
            ],
            "main_scope": "__main__",
            "identifiers": {
                "__main__.main": {
                    "decorators": [],
                    "pc": 0,
                    "type": "function"
                },
                "__main__.fib": {
                    "decorators": [],
                    "pc": 4,
                    "type": "function"
                }
            }
        }
    `)

	stringToFelt := func(bytecode string) *f.Element {
		felt, err := new(f.Element).SetString(bytecode)
		if err != nil {
			panic(err)
		}
		return felt
	}

	program, err := LoadCairoZeroProgram(content)
	require.NoError(t, err)

	require.Equal(t, &Program{
		Bytecode: []*f.Element{
			stringToFelt("0x01"),
			stringToFelt("0x02"),
			stringToFelt("0x03"),
			stringToFelt("0x04"),
		},
		Entrypoints: map[string]uint64{
			"main": 0,
			"fib":  4,
		},
	},
		program,
	)
}
