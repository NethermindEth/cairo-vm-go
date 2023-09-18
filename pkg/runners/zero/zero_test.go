package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
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

func TestSimpleProgram(t *testing.T) {
	program := createDefaultProgram(`
        [ap] = 2, ap++;
        [ap] = 3, ap++;
        [ap] = 4, ap++;
        [ap] = 4;
        [ap - 1] = [ap];
        ret;
    `)

	runner, err := NewRunner(program, false)
	require.NoError(t, err)

	endPc, err := runner.InitializeMainEntrypoint()
	require.NoError(t, err)

	require.Equal(t, uint64(len(program.Bytecode)), endPc)

	err = runner.RunUntilPc(endPc)
	require.NoError(t, err)

	executionSegment := runner.segments()[VM.ExecutionSegment]

	assert.Equal(
		t,
		createSegment(
			// return fp
			memory.NewMemoryAddress(2, 0),
			// next pc
			len(program.Bytecode),
			2,
			3,
			4,
			4,
		),
		executionSegment,
	)

	assert.Equal(t, uint64(5), runner.vm.Context.Ap)
	assert.Equal(t, uint64(0), runner.vm.Context.Fp)
	assert.Equal(t, uint64(len(program.Bytecode)), runner.vm.Context.Pc)
}

func createSegment(values ...any) *memory.Segment {
	data := make([]memory.Cell, len(values))
	for i := range values {
		if values[i] == nil {
			values[i] = memory.Cell{Value: nil, Accessed: false}
		} else {
			memVal, err := memory.MemoryValueFromAny(values[i])
			if err != nil {
				panic(err)
			}
			data[i] = memory.Cell{Value: memVal, Accessed: true}
		}
	}
	return &memory.Segment{
		Data: data,
	}
}

func createDefaultProgram(code string) *Program {
	bytecode, err := assembler.CasmToBytecode(code)
	if err != nil {
		panic(err)
	}

	program := Program{
		Bytecode: bytecode,
		Entrypoints: map[string]uint64{
			"main": 0,
		},
	}

	return &program
}
