package runner

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"

	zero "github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
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

	stringToFelt := func(bytecode string) *fp.Element {
		felt, err := new(fp.Element).SetString(bytecode)
		if err != nil {
			panic(err)
		}
		return felt
	}

	cairoZeroJson, err := zero.ZeroProgramFromJSON(content)
	if err != nil {
		panic(err)
	}

	program, err := LoadCairoZeroProgram(cairoZeroJson)
	require.NoError(t, err)

	require.Equal(t, &Program{
		Bytecode: []*fp.Element{
			stringToFelt("0x01"),
			stringToFelt("0x02"),
			stringToFelt("0x03"),
			stringToFelt("0x04"),
		},
		Entrypoints: map[string]uint64{
			"main": 0,
			"fib":  4,
		},
		Labels: map[string]uint64{},
	},
		program,
	)
}
