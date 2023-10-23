package starknet

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompilerVersionParsing(t *testing.T) {
	testData := []byte(`
       {
          "compiler_version": "2.1.0"
       }
    `)
	starknet, err := StarknetProgramFromJSON(testData)
	require.NoError(t, err)
	assert.Equal(t, "2.1.0", starknet.CompilerVersion)
}

func TestByteCodeParsing(t *testing.T) {
	testData := []byte(`
       {
          "bytecode": [
            "0xa0680017fff8000",
            "0x7",
            "0x482680017ffa8000",
            "0x100000000000000000000000000000000"
          ]
       }
    `)
	starknet, err := StarknetProgramFromJSON(testData)
	require.NoError(t, err)
	assert.Len(t, starknet.Bytecode, 4)
	assert.Equal(t, "482680017ffa8000", starknet.Bytecode[2].Text(16))
}

func TestEmptyEntryPointTypeParsing(t *testing.T) {
	testData := []byte(`
       {
          "entry_points_by_selector": {
                "EXTERNAL": [],
                "L1_HANDLER": [],
                "CONSTRUCTOR": []
          }       
      }
    `)
	starknet, err := StarknetProgramFromJSON(testData)
	require.NoError(t, err)

	entryPoints := starknet.EntryPoints
	assert.Empty(t, entryPoints.External)
	assert.Empty(t, entryPoints.L1Handler)
	assert.Empty(t, entryPoints.Constructor)
}

func TestEntryPointInfoParsing(t *testing.T) {
	testData := []byte(`
       {
          "entry_points_by_type": {
                "EXTERNAL": [
                    {
                        "selector": "0xabcdef0123456789",
                        "offset": 14,
                        "builtins": [
                            "output",
                            "range_check",
                            "pedersen",
                            "ecdsa",
                            "keccak",
                            "bitwise",
                            "ec_op",
                            "poseidon",
                            "segment_arena"
                        ]
                    }
                ],
                "L1_HANDLER": [],
                "CONSTRUCTOR": []
          }       
      }
    `)
	starknet, err := StarknetProgramFromJSON(testData)
	require.NoError(t, err)

	entryPoints := starknet.EntryPoints
	assert.Len(t, entryPoints.External, 1)

	entryPointInfo := entryPoints.External[0]
	assert.Equal(t, entryPointInfo.Selector.Text(16), "abcdef0123456789")
	assert.Equal(t, entryPointInfo.Offset.Text(16), "e")

	assert.Len(t, entryPointInfo.Builtins, 9)
	for i := 0; i < 9; i++ {
		assert.Equal(t, Builtin(i+1), entryPointInfo.Builtins[i])
	}
}

func TestHintsParsing(t *testing.T) {
	v := validator.New()

	testData := []byte(`
       {
            "hints": [
                [
                    2,
                    [
                        {
                            "TestLessThanOrEqual": {
                                "lhs": { "Immediate": "0x95ec" },
                                "rhs": { "Deref": { "register": "FP", "offset": -6 } },
                                "dst": { "register": "AP", "offset": 2 }
                            }
                        }
                    ]
                ],
                [33, [{ "AllocSegment": { "dst": { "register": "AP", "offset": 0 } } }]],
                [
                    52,
                    [
                        {
                            "TestLessThanOrEqual": {
                                "lhs": { "Immediate": "0x1" },
                                "rhs": { "Deref": { "register": "AP", "offset": -13 } },
                                "dst": { "register": "AP", "offset": 1 }
                            }
                        }
                    ]
                ],
                [75, [{ "AllocSegment": { "dst": { "register": "AP", "offset": 1 } } }]],
                [111, [{ "AllocSegment": { "dst": { "register": "AP", "offset": 2 } } }]]
            ]
        }
    `)
	starknet, err := StarknetProgramFromJSON(testData)
	require.NoError(t, err)

	hints := starknet.Hints
	assert.Len(t, hints, 5)

	hint := hints[0].Hints[0]
	assert.Equal(t, hint.Name, TestLessThanOrEqualName)
	_, ok := hint.Args.(*TestLessThanOrEqual)
	assert.True(t, ok)

	assert.NoError(t, v.Struct(starknet))
}

func TestInvalidBuiltin(t *testing.T) {
	testData := []byte(`
       {
          "entry_points_by_type": {
                "EXTERNAL": [
                    {
                        "selector": "0xabcdef0123456789",
                        "offset": 14,
                        "builtins": [
                            "pedrsen",
                        ]
                    }
                ],
                "L1_HANDLER": [],
                "CONSTRUCTOR": []
          }       
      }
    `)
	_, err := StarknetProgramFromJSON(testData)
	assert.Error(t, err)
}
