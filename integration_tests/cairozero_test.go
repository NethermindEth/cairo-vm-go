package integrationtests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCairoZeroFiles(t *testing.T) {
	root := "./cairo_files/"
	testFiles, err := os.ReadDir(root)
	require.NoError(t, err)

	// filter is for debugging purposes
	filter := ""

	for _, dirEntry := range testFiles {
		if dirEntry.IsDir() || isGeneratedFile(dirEntry.Name()) {
			continue
		}

		path := filepath.Join(root, dirEntry.Name())

		if !strings.Contains(path, filter) {
			continue
		}
		t.Logf("testing: %s\n", path)

		compiledOutput, err := compileZeroCode(path)
		if err != nil {
			t.Error(err)
			continue
		}

		outputMode := false
		pyResult, err := runPythonVm(dirEntry.Name(), compiledOutput, outputMode)
		if err != nil {
			t.Error(err)
			continue
		}

		result, err := runVm(compiledOutput, outputMode)
		if err != nil {
			t.Error(err)
			continue
		}

		pyTrace, pyMemory, err := decodeProof(pyResult.TraceFile, pyResult.MemoryFile)
		if err != nil {
			t.Error(err)
			continue
		}

		trace, memory, err := decodeProof(result.TraceFile, result.MemoryFile)
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

func TestCairoZeroOutputFiles(t *testing.T) {
	root := "./output_tests/"

	testFiles, err := os.ReadDir(root)
	require.NoError(t, err)

	runTest := func(t *testing.T, path string) {
		compiledFilename, err := compileZeroCode(path)
		if err != nil {
			t.Fatalf("compiling Cairo0: %v", err)
		}

		outputMode := true
		expectedResult, err := runPythonVm(filepath.Base(path), compiledFilename, outputMode)
		if err != nil {
			t.Fatal(err)
		}

		actualResult, err := runVm(compiledFilename, outputMode)
		if err != nil {
			t.Fatal(err)
		}

		expectedLines := extractOutput([]byte(expectedResult.Output))
		actualLines := extractOutput([]byte(actualResult.Output))
		assert.Equal(t, expectedLines, actualLines)
	}

	for _, f := range testFiles {
		if f.IsDir() || isGeneratedFile(f.Name()) {
			continue
		}

		name := filepath.Base(f.Name())
		name = strings.TrimSuffix(name, filepath.Ext(name))
		t.Run(name, func(t *testing.T) {
			runTest(t, filepath.Join(root, f.Name()))
		})
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

type vmRunResult struct {
	TraceFile  string
	MemoryFile string
	Output     []byte
}

// given a path to a compiled cairo zero file, execute it using the
// python vm and returns the trace and memory files location
func runPythonVm(testFilename, path string, outputMode bool) (vmRunResult, error) {
	var result vmRunResult
	if !outputMode {
		result.TraceFile = swapExtenstion(path, pyTraceSuffix)
		result.MemoryFile = swapExtenstion(path, pyMemorySuffix)
	}

	args := []string{
		"--program",
		path,
	}

	if outputMode {
		args = append(args, "--print_output")
	} else {
		// Proof mode.
		args = append(args,
			"--proof_mode",
			"--trace_file",
			result.TraceFile,
			"--memory_file",
			result.MemoryFile)
	}

	// If any other layouts are needed, add the suffix checks here.
	// The convention would be: ".$layout.cairo"
	// A file without this suffix will use the default ("plain") layout.
	var layout string
	if strings.HasSuffix(testFilename, ".small.cairo") {
		layout = "small"
	}
	if layout == "" && outputMode {
		layout = "small"
	}
	if layout != "" {
		args = append(args, "--layout", layout)
	}

	cmd := exec.Command("cairo-run", args...)

	res, err := cmd.CombinedOutput()
	if outputMode {
		result.Output = res
	}
	if err != nil {
		return result, fmt.Errorf(
			"cairo-run %s: %w\n%s", path, err, string(res),
		)
	}

	return result, nil
}

// given a path to a compiled cairo zero file, execute
// it using our vm
func runVm(path string, outputMode bool) (vmRunResult, error) {
	var result vmRunResult
	if !outputMode {
		result.TraceFile = swapExtenstion(path, traceSuffix)
		result.MemoryFile = swapExtenstion(path, memorySuffix)
	}

	cmd := exec.Command(
		"../bin/cairo-vm",
		"run",
		"--proofmode",
		"--tracefile",
		result.TraceFile,
		"--memoryfile",
		result.MemoryFile,
		path,
	)

	res, err := cmd.CombinedOutput()
	if outputMode {
		result.Output = res
	}
	if err != nil {
		return result, fmt.Errorf(
			"cairo-vm run %s: %w\n%s", path, err, string(res),
		)
	}

	return result, nil

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

func extractOutput(out []byte) []string {
	// Return the output in form of string lines,
	// since then it will be easier for a diff tool
	// to report the mismatching delta.
	const header = "Program output:\n"
	start := bytes.Index(out, []byte(header))
	if start == -1 {
		return nil
	}
	out = out[start+len(header):]
	lines := strings.Split(string(out), "\n")

	// cairo-run produces one extra newline after the output.
	// Remove all empty trailing lines.
	for lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
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

	_, err = runVm(compiledOutput, false)
	require.ErrorContains(t, err, "check write: 2**128 <")

	clean("./builtin_tests/")
}

func TestBitwise(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/bitwise_builtin_test.cairo")
	require.NoError(t, err)

	_, err = runVm(compiledOutput, false)
	require.NoError(t, err)

	clean("./builtin_tests/")
}

func TestPedersen(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/pedersen_test.cairo")
	require.NoError(t, err)

	res, err := runVm(compiledOutput, true)
	require.NoError(t, err)
	require.Contains(t, string(res.Output), "Program output:\n\t2089986280348253421170679821480865132823066470938446095505822317253594081284")

	clean("./builtin_tests/")
}

func TestECDSA(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/ecdsa_test.cairo")
	require.NoError(t, err)

	_, err = runVm(compiledOutput, false)
	//Note: This fails because no addSiganture hint
	require.Error(t, err)

	clean("./builtin_tests/")
}

func TestEcOp(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/ecop.cairo")
	require.NoError(t, err)

	_, err = runVm(compiledOutput, false)
	// todo(rodro): This test is failing due to the lack of hint processing. It should be address soon
	require.Error(t, err)

	clean("./builtin_tests/")
}

func TestKeccak(t *testing.T) {
	compiledOutput, err := compileZeroCode("./builtin_tests/keccak_test.cairo")
	require.NoError(t, err)

	res, err := runVm(compiledOutput, true)
	require.NoError(t, err)
	require.Contains(t, string(res.Output), "Program output:\n\t1304102964824333531548398680304964155037696012322029952943772\n\t688749063493959345342507274897412933692859993314608487848187\n\t986714560881445649520443980361539218531403996118322524237197\n\t1184757872753521629808292433475729390634371625298664050186717\n\t719230200744669084408849842242045083289669818920073250264351\n\t1543031433416778513637578850638598357854418012971636697855068\n\t63644822371671650271181212513090078620238279557402571802224\n\t879446821229338092940381117330194802032344024906379963157761\n")

	clean("./builtin_tests/")
}
