package zero

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleProgram(t *testing.T) {
	program := createDefaultProgram(`
        [ap] = 2, ap++;
        [ap] = 3, ap++;
        [ap] = 4, ap++;
        [ap] = 4;
        [ap - 1] = [ap];
        ret;
    `)

	runner, err := NewRunner(program, false, math.MaxUint64)
	require.NoError(t, err)

	endPc, err := runner.InitializeMainEntrypoint()
	require.NoError(t, err)

	expectedPc := memory.NewMemoryAddress(3, 0)

	require.Equal(t, expectedPc, endPc)

	err = runner.RunUntilPc(endPc)
	require.NoError(t, err)

	executionSegment := runner.segments()[VM.ExecutionSegment]

	assert.Equal(
		t,
		createSegment(
			// return fp
			memory.NewMemoryAddress(2, 0),
			// next pc
			expectedPc,
			2,
			3,
			4,
			4,
		),
		trimmedSegment(executionSegment),
	)

	assert.Equal(t, uint64(5), runner.vm.Context.Ap)
	assert.Equal(t, uint64(0), runner.vm.Context.Fp)
	assert.Equal(t, expectedPc, runner.vm.Context.Pc)
}

func TestStepLimitExceeded(t *testing.T) {
	program := createDefaultProgram(`
        [ap] = 2;
        [ap + 1] = 3;
        [ap + 2] = 5;
        [ap + 3] = 7;
        [ap + 4] = 11;
        [ap + 5] = 13;
        ret;
    `)

	runner, err := NewRunner(program, false, 3)
	require.NoError(t, err)

	endPc, err := runner.InitializeMainEntrypoint()
	require.NoError(t, err)

	expectedPc := memory.NewMemoryAddress(3, 0)
	require.Equal(t, expectedPc, endPc)

	err = runner.RunUntilPc(endPc)
	require.ErrorContains(t, err, "step limit exceeded")

	executionSegment := runner.segments()[VM.ExecutionSegment]

	assert.Equal(
		t,
		createSegment(
			// return fp
			memory.NewMemoryAddress(2, 0),
			// next pc
			expectedPc,
			2,
			3,
			5,
		),
		executionSegment,
	)

	// when running on non proof mode, the first to elements
	// are dummy values. So ap and fp starts at 2
	assert.Equal(t, uint64(2), runner.vm.Context.Ap)
	assert.Equal(t, uint64(2), runner.vm.Context.Fp)
	// the fourth instruction starts at 0:6 because all previous one have size 2
	assert.Equal(t, memory.NewMemoryAddress(0, 6), runner.vm.Context.Pc)
	// step limit exceeded
	assert.Equal(t, uint64(3), runner.steps())
}

func TestStepLimitExceededProofMode(t *testing.T) {
	program := createDefaultProgram(`
        [ap] = 2;
        [ap + 1] = 3;
        [ap + 2] = 5;
        [ap + 3] = 7;
        [ap + 4] = 11;
        [ap + 5] = 13;
        jmp rel 0;
    `)
	// properties required by proofmode
	program.Labels = map[string]uint64{
		"__start__": 0,
		"__end__":   uint64(len(program.Bytecode)),
	}

	for _, maxstep := range []uint64{6, 7} {
		t.Logf("Using maxstep: %d\n", maxstep)
		// when maxstep = 6, it fails executing the extra step required by proof mode
		// when maxstep = 7, it fails trying to get the trace to be a power of 2
		runner, err := NewRunner(program, true, uint64(maxstep))
		require.NoError(t, err)

		err = runner.Run()
		require.ErrorContains(t, err, "step limit exceeded")

		executionSegment := runner.segments()[VM.ExecutionSegment]

		assert.Equal(
			t,
			createSegment(
				// return fp
				memory.NewMemoryAddress(
					0,
					uint64(len(program.Bytecode)+2),
				),
				// next pc
				0,
				2,
				3,
				5,
				7,
				11,
				13,
			),
			executionSegment,
		)

		// when running on non proof mode, the first to elements
		// are dummy values. So ap and fp starts at 2
		assert.Equal(t, uint64(2), runner.vm.Context.Ap)
		assert.Equal(t, uint64(2), runner.vm.Context.Fp)
		// it repeats the last instruction at 0:12
		assert.Equal(t, memory.NewMemoryAddress(0, 12), runner.vm.Context.Pc)
		// step limit exceeded
		assert.Equal(t, uint64(maxstep), runner.steps())
	}
}

func TestTraceEncodingDecoding(t *testing.T) {
	trace := []vm.Trace{
		{Ap: 1, Fp: 2, Pc: 3},
		{Ap: 4, Fp: 5, Pc: 6},
		{Ap: 9, Fp: 8, Pc: 7},
	}

	encodedTrace := EncodeTrace(trace)

	expected := make([]byte, len(trace)*3*8)
	// first context
	binary.LittleEndian.PutUint64(expected[0:8], 1)
	binary.LittleEndian.PutUint64(expected[8:16], 2)
	binary.LittleEndian.PutUint64(expected[16:24], 3)
	// second context
	binary.LittleEndian.PutUint64(expected[24:32], 4)
	binary.LittleEndian.PutUint64(expected[32:40], 5)
	binary.LittleEndian.PutUint64(expected[40:48], 6)
	// third context
	binary.LittleEndian.PutUint64(expected[48:56], 9)
	binary.LittleEndian.PutUint64(expected[56:64], 8)
	binary.LittleEndian.PutUint64(expected[64:72], 7)

	// test encoding
	require.Equal(
		t,
		expected,
		encodedTrace,
	)

	// test decoding
	decodedTrace := DecodeTrace(encodedTrace)
	require.Equal(
		t,
		trace,
		decodedTrace,
	)

}

func TestMemoryEncodingDecoding(t *testing.T) {
	memory := []*f.Element{
		new(f.Element).SetUint64(4),
		new(f.Element).SetUint64(15),
		nil,
		nil,
		new(f.Element).SetUint64(8),
		nil,
		new(f.Element).SetUint64(2),
	}

	encodedMemory := EncodeMemory(memory)

	// the array size depends on the ammount of non nil elements
	// it stores (addres, felt) encoded in little endian in a consecutive way
	expected := make([]byte, 4*(8+32))

	//first element
	binary.LittleEndian.PutUint64(expected[0:8], 0)
	f.LittleEndian.PutElement((*[32]byte)(expected[8:40]), *new(f.Element).SetUint64(4))
	//second element
	binary.LittleEndian.PutUint64(expected[40:48], 1)
	f.LittleEndian.PutElement((*[32]byte)(expected[48:80]), *new(f.Element).SetUint64(15))
	//third element
	binary.LittleEndian.PutUint64(expected[80:88], 4)
	f.LittleEndian.PutElement((*[32]byte)(expected[88:120]), *new(f.Element).SetUint64(8))
	//fourth element
	binary.LittleEndian.PutUint64(expected[120:128], 6)
	f.LittleEndian.PutElement((*[32]byte)(expected[128:160]), *new(f.Element).SetUint64(2))

	require.Equal(
		t,
		len(expected),
		len(encodedMemory),
	)
	require.Equal(
		t,
		expected,
		encodedMemory,
	)

	// testing decoding
	decodedMemory := DecodeMemory(encodedMemory)
	require.Equal(
		t,
		memory,
		decodedMemory,
	)
}

func BenchmarkRunnerWithFibonacci(b *testing.B) {
	compiledJson := []byte(`
        {
            "compiler_version": "0.11.0.2",
            "data": [
                "0x40780017fff7fff",
                "0x0",
                "0x1104800180018000",
                "0x4",
                "0x10780017fff7fff",
                "0x0",
                "0x480680017fff8000",
                "0x1",
                "0x480680017fff8000",
                "0x1",
                "0x480680017fff8000",
                "0xf4240",
                "0x1104800180018000",
                "0x3",
                "0x208b7fff7fff7ffe",
                "0x20780017fff7ffd",
                "0x4",
                "0x480a7ffc7fff8000",
                "0x208b7fff7fff7ffe",
                "0x482a7ffc7ffb8000",
                "0x480a7ffc7fff8000",
                "0x48127ffe7fff8000",
                "0x482680017ffd8000",
                "0x800000000000011000000000000000000000000000000000000000000000000",
                "0x1104800180018000",
                "0x800000000000010fffffffffffffffffffffffffffffffffffffffffffffff8",
                "0x208b7fff7fff7ffe"
            ],
            "identifiers": {
                "__main__.__end__": {
                    "pc": 4,
                    "type": "label"
                },
                "__main__.__start__": {
                    "pc": 0,
                    "type": "label"
                },
                "__main__.fib": {
                    "decorators": [],
                    "pc": 15,
                    "type": "function"
                },
                "__main__.fib.Args": {
                    "full_name": "__main__.fib.Args",
                    "members": {
                        "first_element": {
                            "cairo_type": "felt",
                            "offset": 0
                        },
                        "n": {
                            "cairo_type": "felt",
                            "offset": 2
                        },
                        "second_element": {
                            "cairo_type": "felt",
                            "offset": 1
                        }
                    },
                    "size": 3,
                    "type": "struct"
                },
                "__main__.fib.ImplicitArgs": {
                    "full_name": "__main__.fib.ImplicitArgs",
                    "members": {},
                    "size": 0,
                    "type": "struct"
                },
                "__main__.fib.Return": {
                    "cairo_type": "(res: felt)",
                    "type": "type_definition"
                },
                "__main__.fib.SIZEOF_LOCALS": {
                    "type": "const",
                    "value": 0
                },
                "__main__.fib.first_element": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fib.first_element",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 4,
                                "offset": 0
                            },
                            "pc": 15,
                            "value": "[cast(fp + (-5), felt*)]"
                        }
                    ],
                    "type": "reference"
                },
                "__main__.fib.n": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fib.n",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 4,
                                "offset": 0
                            },
                            "pc": 15,
                            "value": "[cast(fp + (-3), felt*)]"
                        }
                    ],
                    "type": "reference"
                },
                "__main__.fib.second_element": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fib.second_element",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 4,
                                "offset": 0
                            },
                            "pc": 15,
                            "value": "[cast(fp + (-4), felt*)]"
                        }
                    ],
                    "type": "reference"
                },
                "__main__.fib.y": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fib.y",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 4,
                                "offset": 1
                            },
                            "pc": 20,
                            "value": "[cast(ap + (-1), felt*)]"
                        }
                    ],
                    "type": "reference"
                },
                "__main__.main": {
                    "decorators": [],
                    "pc": 6,
                    "type": "function"
                },
                "__main__.main.Args": {
                    "full_name": "__main__.main.Args",
                    "members": {},
                    "size": 0,
                    "type": "struct"
                },
                "__main__.main.ImplicitArgs": {
                    "full_name": "__main__.main.ImplicitArgs",
                    "members": {},
                    "size": 0,
                    "type": "struct"
                },
                "__main__.main.Return": {
                    "cairo_type": "()",
                    "type": "type_definition"
                },
                "__main__.main.SIZEOF_LOCALS": {
                    "type": "const",
                    "value": 0
                }
            },
            "main_scope": "__main__",
            "prime": "0x800000000000011000000000000000000000000000000000000000000000001",
            "reference_manager": {
                "references": [
                    {
                        "ap_tracking_data": {
                            "group": 4,
                            "offset": 0
                        },
                        "pc": 15,
                        "value": "[cast(fp + (-5), felt*)]"
                    },
                    {
                        "ap_tracking_data": {
                            "group": 4,
                            "offset": 0
                        },
                        "pc": 15,
                        "value": "[cast(fp + (-4), felt*)]"
                    },
                    {
                        "ap_tracking_data": {
                            "group": 4,
                            "offset": 0
                        },
                        "pc": 15,
                        "value": "[cast(fp + (-3), felt*)]"
                    },
                    {
                        "ap_tracking_data": {
                            "group": 4,
                            "offset": 1
                        },
                        "pc": 20,
                        "value": "[cast(ap + (-1), felt*)]"
                    }
                ]
            }
        }
    `)
	for i := 0; i < b.N; i++ {
		program, err := LoadCairoZeroProgram(compiledJson)
		if err != nil {
			panic(err)
		}

		runner, err := NewRunner(program, true)
		if err != nil {
			panic(err)
		}

		err = runner.Run()
		if err != nil {
			panic(err)
		}
	}
}

func createSegment(values ...any) *memory.Segment {
	data := make([]*memory.Cell, len(values))
	for i := range values {
		if values[i] != nil {
			memVal, err := memory.MemoryValueFromAny(values[i])
			if err != nil {
				panic(err)
			}
			data[i] = &memory.Cell{Value: memVal, Accessed: true}
		}
	}
	return &memory.Segment{
		Data:      data,
		LastIndex: len(data) - 1,
	}
}

// modifies a segment in place to reduce its real length to
// its effective lenth. It returns the same segment
func trimmedSegment(segment *memory.Segment) *memory.Segment {
	segment.Data = segment.Data[0:segment.Len()]
	return segment
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
