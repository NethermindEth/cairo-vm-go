package vm

import (
	"github.com/stretchr/testify/require"
	"testing"
    f "github.com/NethermindEth/juno/core/felt"
)

func TestValidPrime(t *testing.T) {
	testData := []byte(`
       {
          "prime": "0x800000000000011000000000000000000000000000000000000000000000001",
          "compiler_version": "2.1.0",
          "bytecode": [
            "0xa0680017fff8000",
            "0x7",
            "0x482680017ffa8000",
            "0x100000000000000000000000000000000"
          ]
       }
    `)
	_, err := ParseProgram(testData)
	require.NoError(t, err)

}

func TestInvalidPrime(t *testing.T) {

}
