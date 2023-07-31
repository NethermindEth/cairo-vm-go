package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidPrime(t *testing.T) {
	testData := []byte(`
       {
          "compiler_version": "2.1.0",
          "bytecode": [
            "0xa0680017fff8000",
            "0x7",
            "0x482680017ffa8000",
            "0x100000000000000000000000000000000"
          ]
       }
    `)
	program, err := ProgramFromJSON(testData)
	require.NoError(t, err)
	assert.Equal(t, "2.1.0", program.CompilerVersion)
	assert.Len(t, program.Bytecode, 4)
	assert.Equal(t, "0x482680017ffa8000", program.Bytecode[2].String())

}

func TestInvalidPrime(t *testing.T) {

}
