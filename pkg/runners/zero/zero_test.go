package zero

import (
	"math"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
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

	expectedPc := memory.MemoryAddress{SegmentIndex: 3, Offset: 0}

	require.Equal(t, expectedPc, endPc)

	err = runner.RunUntilPc(&endPc)
	require.NoError(t, err)

	executionSegment := runner.vm.Memory.Segments[vm.ExecutionSegment]

	assert.Equal(
		t,
		createSegment(
			// return fp
			&memory.MemoryAddress{SegmentIndex: 2, Offset: 0},
			// next pc
			&expectedPc,
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

	expectedPc := memory.MemoryAddress{SegmentIndex: 3, Offset: 0}
	require.Equal(t, expectedPc, endPc)

	err = runner.RunUntilPc(&endPc)
	require.ErrorContains(t, err, "step limit exceeded")

	executionSegment := runner.vm.Memory.Segments[vm.ExecutionSegment]

	assert.Equal(
		t,
		createSegment(
			// return fp
			&memory.MemoryAddress{SegmentIndex: 2, Offset: 0},
			// next pc
			&expectedPc,
			2,
			3,
			5,
		),
		trimmedSegment(executionSegment),
	)

	// when running on non proof mode, the first to elements
	// are dummy values. So ap and fp starts at 2
	assert.Equal(t, uint64(2), runner.vm.Context.Ap)
	assert.Equal(t, uint64(2), runner.vm.Context.Fp)
	// the fourth instruction starts at 0:6 because all previous one have size 2
	assert.Equal(t, memory.MemoryAddress{SegmentIndex: 0, Offset: 6}, runner.vm.Context.Pc)
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

		executionSegment := runner.vm.Memory.Segments[vm.ExecutionSegment]

		assert.Equal(
			t,
			createSegment(
				// return fp
				&memory.MemoryAddress{SegmentIndex: 0, Offset: uint64(len(program.Bytecode) + 2)},
				// next pc
				0,
				2,
				3,
				5,
				7,
				11,
				13,
			),
			trimmedSegment(executionSegment),
		)

		// when running on non proof mode, the first to elements
		// are dummy values. So ap and fp starts at 2
		assert.Equal(t, uint64(2), runner.vm.Context.Ap)
		assert.Equal(t, uint64(2), runner.vm.Context.Fp)
		// it repeats the last instruction at 0:12
		assert.Equal(t, memory.MemoryAddress{SegmentIndex: 0, Offset: 12}, runner.vm.Context.Pc)
		// step limit exceeded
		assert.Equal(t, uint64(maxstep), runner.steps())
	}
}

func createSegment(values ...any) *memory.Segment {
	data := make([]memory.MemoryValue, len(values))
	for i := range values {
		if values[i] != nil {
			var err error
			data[i], err = memory.MemoryValueFromAny(values[i])
			if err != nil {
				panic(err)
			}
		}
	}
	s := &memory.Segment{
		Data:      data,
		LastIndex: len(data) - 1,
	}
	s.WithBuiltinRunner(&memory.NoBuiltin{})
	return s
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
