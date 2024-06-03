package integrationtests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Filter struct {
	filters []string
}

func (f *Filter) init() {
	filtersRaw := os.Getenv("INTEGRATION_TESTS_FILTERS")
	if filtersRaw == "" {
		_ = godotenv.Load("./.env")
		filtersRaw = os.Getenv("INTEGRATION_TESTS_FILTERS")
	}
	filters := strings.Split(filtersRaw, ",")
	for _, filter := range filters {
		trimmed := strings.TrimSpace(filter)
		if trimmed != "" {
			f.filters = append(f.filters, trimmed)
		}
	}
}

func (f *Filter) filtered(testFile string) bool {
	if len(f.filters) == 0 {
		return true
	}

	for _, filter := range f.filters {
		if strings.Contains(testFile, filter) {
			return true
		}
	}

	return false
}

func TestCairoZeroFiles(t *testing.T) {
	root := "./cairo_files/"
	testFiles, err := os.ReadDir(root)
	require.NoError(t, err)

	// filter is for debugging purposes
	filter := Filter{}
	filter.init()

	for _, dirEntry := range testFiles {
		if dirEntry.IsDir() || isGeneratedFile(dirEntry.Name()) {
			continue
		}

		path := filepath.Join(root, dirEntry.Name())

		if !filter.filtered(dirEntry.Name()) {
			continue
		}

		t.Logf("testing: %s\n", path)

		compiledOutput, err := compileZeroCode(path)
		if err != nil {
			t.Error(err)
			continue
		}

		pyTraceFile, pyMemoryFile, err := runPythonVm(dirEntry.Name(), compiledOutput)
		if err != nil {
			t.Error(err)
			continue
		}

		traceFile, memoryFile, _, err := runVm(compiledOutput)
		if err != nil {
			t.Error(err)
			continue
		}

		pyTrace, pyMemory, err := decodeProof(pyTraceFile, pyMemoryFile)
		if err != nil {
			t.Error(err)
			continue
		}

		trace, memory, err := decodeProof(traceFile, memoryFile)
		if err != nil {
			t.Error(err)
			continue
		}

		if !assert.Equal(t, pyTrace, trace) {
			t.Logf("pytrace:\n%s\n", traceRepr(pyTrace))
			t.Logf("trace:\n%s\n", traceRepr(trace))
		}
		if !assert.Equal(t, pyMemory, memory) {
			t.Logf("pymemory;\n%s\n", memoryRepr(pyMemory))
			t.Logf("memory;\n%s\n", memoryRepr(memory))
		}
	}

	clean(root)
}

const (
	compiledSuffix = "_compiled.json"
	pyTraceSuffix  = "_py_trace"
	pyMemorySuffix = "_py_memory"
	traceSuffix    = "_trace"
	memorySuffix   = "_memory"
)

// given a path to a cairo zero file, it compiles it
// and return the compilation path
func compileZeroCode(path string) (string, error) {
	if filepath.Ext(path) != ".cairo" {
		return "", fmt.Errorf("compiling non cairo file: %s", path)
	}
	compiledOutput := swapExtenstion(path, compiledSuffix)

	cmd := exec.Command(
		"cairo-compile",
		path,
		"--proof_mode",
		"--no_debug_info",
		"--output",
		compiledOutput,
	)

	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"cairo-compile %s: %w\n%s", path, err, string(res),
		)
	}

	return compiledOutput, nil
}

// given a path to a compiled cairo zero file, execute it using the
// python vm and returns the trace and memory files location
func runPythonVm(testFilename, path string) (string, string, error) {
	traceOutput := swapExtenstion(path, pyTraceSuffix)
	memoryOutput := swapExtenstion(path, pyMemorySuffix)

	args := []string{
		"--program",
		path,
		"--proof_mode",
		"--trace_file",
		traceOutput,
		"--memory_file",
		memoryOutput,
	}

	// If any other layouts are needed, add the suffix checks here.
	// The convention would be: ".$layout.cairo"
	// A file without this suffix will use the default ("plain") layout.
	if strings.HasSuffix(testFilename, ".small.cairo") {
		args = append(args, "--layout", "small")
	}

	cmd := exec.Command("cairo-run", args...)

	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf(
			"cairo-run %s: %w\n%s", path, err, string(res),
		)
	}

	return traceOutput, memoryOutput, nil
}

// given a path to a compiled cairo zero file, execute
// it using our vm
func runVm(path string) (string, string, string, error) {
	traceOutput := swapExtenstion(path, traceSuffix)
	memoryOutput := swapExtenstion(path, memorySuffix)

	cmd := exec.Command(
		"../bin/cairo-vm",
		"run",
		"--proofmode",
		"--tracefile",
		traceOutput,
		"--memoryfile",
		memoryOutput,
		path,
	)

	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", string(res), fmt.Errorf(
			"cairo-vm run %s: %w\n%s", path, err, string(res),
		)
	}

	return traceOutput, memoryOutput, string(res), nil

}

func decodeProof(traceLocation string, memoryLocation string) ([]vm.Trace, []*fp.Element, error) {
	trace, err := os.ReadFile(traceLocation)
	if err != nil {
		return nil, nil, err
	}
	decodedTrace := vm.DecodeTrace(trace)

	memory, err := os.ReadFile(memoryLocation)
	if err != nil {
		return nil, nil, err
	}
	decodedMemory := vm.DecodeMemory(memory)

	return decodedTrace, decodedMemory, nil
}

func clean(root string) {
	err := filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if isGeneratedFile(path) {
				return os.Remove(path)
			}
			return nil
		},
	)

	if err != nil {
		panic(err)
	}
}

// Given a certain file, it swaps its extension with a new suffix
func swapExtenstion(path string, newSuffix string) string {
	dir, name := filepath.Split(path)
	name = name[:len(name)-len(".cairo")]
	name = fmt.Sprintf("%s%s", name, newSuffix)
	return filepath.Join(dir, name)
}

func isGeneratedFile(path string) bool {
	return strings.HasSuffix(path, compiledSuffix) ||
		strings.HasSuffix(path, pyTraceSuffix) ||
		strings.HasSuffix(path, pyMemorySuffix) ||
		strings.HasSuffix(path, traceSuffix) ||
		strings.HasSuffix(path, memorySuffix)
}

func traceRepr(trace []vm.Trace) string {
	repr := make([]string, len(trace))
	for i, ctx := range trace {
		repr[i] = fmt.Sprintf("{pc: %d, ap: %d, fp: %d}", ctx.Pc, ctx.Ap, ctx.Fp)
	}
	return strings.Join(repr, ", ")
}

func memoryRepr(memory []*fp.Element) string {
	repr := make([]string, len(memory))
	for i, felt := range memory {
		repr[i] = felt.Text(10)
	}
	return strings.Join(repr, ", ")

}

func TestFailingRangeCheck(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/range_check.cairo")
	require.NoError(t, err)

	_, _, _, err = runVm(compiledOutput)
	require.ErrorContains(t, err, "check write: 2**128 <")

	clean("./builtin_tests/")
}

func TestBitwise(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/bitwise_builtin_test.cairo")
	require.NoError(t, err)

	_, _, _, err = runVm(compiledOutput)
	require.NoError(t, err)

	clean("./builtin_tests/")
}

func TestPedersen(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/pedersen_test.cairo")
	require.NoError(t, err)

	_, _, output, err := runVm(compiledOutput)
	require.NoError(t, err)
	require.Contains(t, output, "Program output:\n  2089986280348253421170679821480865132823066470938446095505822317253594081284")

	clean("./builtin_tests/")
}

func TestPoseidon(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/poseidon_test.cairo")
	require.NoError(t, err)

	_, _, output, err := runVm(compiledOutput)
	require.NoError(t, err)
	require.Contains(t, output, "Program output:\n  2089986280348253421170679821480865132823066470938446095505822317253594081284")

	clean("./builtin_tests/")
}

func TestECDSA(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/ecdsa_test.cairo")
	require.NoError(t, err)

	_, _, _, err = runVm(compiledOutput)
	require.NoError(t, err)

	clean("./builtin_tests/")
}

func TestEcOp(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/ecop.cairo")
	require.NoError(t, err)

	_, _, _, err = runVm(compiledOutput)
	// todo(rodro): This test is failing due to the lack of hint processing. It should be address soon
	require.Error(t, err)

	clean("./builtin_tests/")
}

func TestKeccak(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/keccak_test.cairo")
	require.NoError(t, err)

	_, _, output, err := runVm(compiledOutput)
	require.NoError(t, err)
	require.Contains(t, output, "Program output:\n  1304102964824333531548398680304964155037696012322029952943772\n  688749063493959345342507274897412933692859993314608487848187\n  986714560881445649520443980361539218531403996118322524237197\n  1184757872753521629808292433475729390634371625298664050186717\n  719230200744669084408849842242045083289669818920073250264351\n  1543031433416778513637578850638598357854418012971636697855068\n  63644822371671650271181212513090078620238279557402571802224\n  879446821229338092940381117330194802032344024906379963157761\n")

	clean("./builtin_tests/")
}
