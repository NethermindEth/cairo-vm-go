package zero

import (
	"fmt"
	"math"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	sn "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	pedersenhash "github.com/consensys/gnark-crypto/ecc/stark-curve/pedersen-hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleProgram(t *testing.T) {
	program := createProgram(`
        [ap] = 2, ap++;
        [ap] = 3, ap++;
        [ap] = 4, ap++;
        [ap] = 4;
        [ap - 1] = [ap];
        ret;
    `)

	hints := make(map[uint64][]hinter.Hinter)
	runner, err := NewRunner(program, hints, false, math.MaxUint64)
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
	program := createProgram(`
        [ap] = 2;
        [ap + 1] = 3;
        [ap + 2] = 5;
        [ap + 3] = 7;
        [ap + 4] = 11;
        [ap + 5] = 13;
        ret;
    `)

	hints := make(map[uint64][]hinter.Hinter)
	runner, err := NewRunner(program, hints, false, 3)
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
	program := createProgram(`
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
		hints := make(map[uint64][]hinter.Hinter)
		runner, err := NewRunner(program, hints, true, uint64(maxstep))
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

func TestBitwiseBuiltin(t *testing.T) {
	// bitwise segment ptr is located at fp - 3 (fp - 2 and fp - 1 contain initialization vals)
	// We first write 16 and 8 to bitwise. Then we read the bitwise result from &, ^ and |
	runner := createRunner(`
        [ap] = 14, ap++;
        [ap] = 7, ap++;
        [ap - 2] = [[fp - 3]];
        [ap - 1] = [[fp - 3] + 1];

        [ap] = [[fp - 3] + 2];
        [ap + 1] = [[fp - 3] + 3];
        [ap + 2] = [[fp - 3] + 4];

        [ap] = 6;
        [ap + 1] = 9;
        [ap + 2] = 15;
        ret;
    `, sn.Bitwise)

	err := runner.Run()
	require.NoError(t, err)

	bitwise, ok := runner.vm.Memory.FindSegmentWithBuiltin("bitwise")
	require.True(t, ok)

	requireEqualSegments(t, createSegment(14, 7, 6, 9, 15), bitwise)
}

func TestBitwiseBuiltinError(t *testing.T) {
	// inferring first write to cell
	runner := createRunner(`
	    [ap] = [[fp - 3]];
	    ret;
	`, sn.Bitwise)

	err := runner.Run()
	require.ErrorContains(t, err, "cannot infer value")

	// inferring second write to cell
	runner = createRunner(`
	    [ap] = [[fp - 3] + 1];
	    ret;
	`, sn.Bitwise)
	err = runner.Run()
	require.ErrorContains(t, err, "cannot infer value")

	// trying to infer without writing before
	runner = createRunner(`
        [ap] = [[fp - 3] + 2];
        ret;
    `, sn.Bitwise)

	err = runner.Run()
	require.ErrorContains(t, err, "input value at offset 0 is unknown")
}

func TestOutputBuiltin(t *testing.T) {
	// Output builtin is located at fp - 3
	runner := createRunner(`
        [ap] = 5;
        [ap] = [[fp - 3]];
        [ap + 1] = 7;
        [ap + 1] = [[fp - 3] + 1];
        ret;
    `, sn.Output)
	err := runner.Run()
	require.NoError(t, err)

	output := runner.Output()

	val1 := fp.NewElement(5)
	val2 := fp.NewElement(7)
	require.Equal(t, []*fp.Element{&val1, &val2}, output)
}

func TestPedersenBuiltin(t *testing.T) {
	val1 := fp.NewElement(5)
	val2 := fp.NewElement(7)
	val3 := pedersenhash.Pedersen(&val1, &val2)

	// pedersen builtin is located at fp - 3
	// we first write val1 and val2 and then check the infered value is val3
	code := fmt.Sprintf(`
        [ap] = %s;
        [ap] = [[fp - 3]];

        [ap + 1] = %s;
        [ap + 1] = [[fp - 3] + 1];

        [ap + 2] = [[fp - 3] + 2];
        [ap + 2] = %s;
        ret;
    `, val1.Text(10), val2.Text(10), val3.Text(10))

	runner := createRunner(code, sn.Pedersen)
	err := runner.Run()
	require.NoError(t, err)

	pedersen, ok := runner.vm.Memory.FindSegmentWithBuiltin("pedersen")
	require.True(t, ok)
	requireEqualSegments(t, createSegment(&val1, &val2, &val3), pedersen)
}

func TestPedersenBuiltinError(t *testing.T) {
	runner := createRunner(`
        [ap] = [[fp - 3]];
        ret;
    `, sn.Pedersen)
	err := runner.Run()
	require.ErrorContains(t, err, "cannot infer value")

	runner = createRunner(`
        [ap] = [[fp - 3] + 2];
        ret;
    `, sn.Pedersen)
	err = runner.Run()
	require.ErrorContains(t, err, "input value at offset 0 is unknown")
}

func TestRangeCheckBuiltin(t *testing.T) {
	// range check is located at fp - 3 (fp - 2 and fp - 1 contain initialization vals)
	// we write 5 and 2**128 - 1 to range check
	// no error should come from this
	runner := createRunner(`
        [ap] = 5;
        [ap] = [[fp - 3]];
        [ap + 1] = 0xffffffffffffffffffffffffffffffff;
        [ap + 1] = [[fp - 3] + 1];
        ret;
    `, sn.RangeCheck)

	err := runner.Run()
	require.NoError(t, err)

	rangeCheck, ok := runner.vm.Memory.FindSegmentWithBuiltin("range_check")
	require.True(t, ok)

	felt := &fp.Element{}
	felt, err = felt.SetString("0xffffffffffffffffffffffffffffffff")
	require.NoError(t, err)

	requireEqualSegments(t, createSegment(5, felt), rangeCheck)
}

func TestRangeCheckBuiltinError(t *testing.T) {
	// first test fails due to out of bound check
	runner := createRunner(`
        [ap] = 0x100000000000000000000000000000000;
        [ap] = [[fp - 3]];
        ret;
    `, sn.RangeCheck)

	err := runner.Run()
	require.ErrorContains(t, err, "check write: 2**128 <")

	// second test fails due to reading unknown value
	runner = createRunner(`
        [ap] = [[fp - 3]];
        ret;
    `, sn.RangeCheck)

	err = runner.Run()
	require.ErrorContains(t, err, "cannot infer value")
}

func TestEcOpBuiltin(t *testing.T) {
	// first, store P.x, P.y, Q.x, Q.y and m in the data segment
	// then store them the EcOp builtin segment
	// infer the values, effectively calculating EcOp
	// assert the values are correct
	runner := createRunner(`
        [ap] = 0x6a4beaef5a93425b973179cdba0c9d42f30e01a5f1e2db73da0884b8d6756fc;
        [ap + 1] = 0x72565ec81bc09ff53fbfad99324a92aa5b39fb58267e395e8abe36290ebf24f;
        [ap + 2] = 0x654fd7e67a123dd13868093b3b7777f1ffef596c2e324f25ceaf9146698482c;
        [ap + 3] = 0x4fad269cbf860980e38768fe9cb6b0b9ab03ee3fe84cfde2eccce597c874fd8;
        [ap + 4] = 34;

        [ap] = [[fp - 3]];
        [ap + 1] = [[fp - 3] + 1];
        [ap + 2] = [[fp - 3] + 2];
        [ap + 3] = [[fp - 3] + 3];
        [ap + 4] = [[fp - 3] + 4];

        [ap + 5] = [[fp - 3] + 5];
        [ap + 6] = [[fp - 3] + 6];

        [ap + 5] = 108925483682366235368969256555281508851459278989259552980345066351008608800;
        [ap + 6] = 1592365885972480102953613056006596671718206128324372995731808913669237079419;
        ret;
    `, sn.ECOP)

	err := runner.Run()
	require.NoError(t, err)
}

func createRunner(code string, builtins ...sn.Builtin) ZeroRunner {
	program := createProgramWithBuiltins(code, builtins...)

	hints := make(map[uint64][]hinter.Hinter)
	runner, err := NewRunner(program, hints, false, math.MaxUint64)
	if err != nil {
		panic(err)
	}
	return runner

}

// utility to create segments easier
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

// compare two segments ignoring builtins
func requireEqualSegments(t *testing.T, expected, result *memory.Segment) {
	result = trimmedSegment(result)

	t.Log(expected)
	t.Log(result)

	assert.Equal(t, expected.LastIndex, result.LastIndex)
	require.Equal(t, expected.Data, result.Data)
}

// modifies a segment in place to reduce its real length to
// its effective lenth. It returns the same segment
func trimmedSegment(segment *memory.Segment) *memory.Segment {
	segment.Data = segment.Data[0:segment.Len()]
	return segment
}

func createProgram(code string) *Program {
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

func createProgramWithBuiltins(code string, builtins ...sn.Builtin) *Program {
	program := createProgram(code)
	program.Builtins = builtins
	return program
}
