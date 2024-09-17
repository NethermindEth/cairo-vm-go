package integrationtests

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"text/tabwriter"
	"time"

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

func runAndTestFile(t *testing.T, path string, name string, benchmarkMap map[string][2]int, benchmark bool, errorExpected bool) {
	t.Logf("testing: %s\n", path)

	compiledOutput, err := compileZeroCode(path)
	if err != nil {
		t.Error(err)
		return
	}

	elapsedPy, pyTraceFile, pyMemoryFile, err := runPythonVm(name, compiledOutput)
	if errorExpected {
		// we let the code go on so that we can check if the go vm also raises an error
		assert.Error(t, err, path)
	} else {
		if err != nil {
			t.Error(err)
			return
		}
	}

	elapsedGo, traceFile, memoryFile, _, err := runVm(compiledOutput)
	if errorExpected {
		assert.Error(t, err, path)
		return
	} else {
		if err != nil {
			t.Error(err)
			return
		}
	}

	if benchmark {
		benchmarkMap[name] = [2]int{int(elapsedPy.Milliseconds()), int(elapsedGo.Milliseconds())}
	}

	pyTrace, pyMemory, err := decodeProof(pyTraceFile, pyMemoryFile)
	if err != nil {
		t.Error(err)
		return
	}

	trace, memory, err := decodeProof(traceFile, memoryFile)
	if err != nil {
		t.Error(err)
		return
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

var zerobench = flag.Bool("zerobench", false, "run integration tests and generate benchmarks file")

func TestCairoZeroFiles(t *testing.T) {
	roots := []string{
		"./cairo_zero_hint_tests/",
		"./cairo_zero_file_tests/",
		"./builtin_tests/",
	}

	// filter is for debugging purposes
	filter := Filter{}
	filter.init()

	benchmarkMap := make(map[string][2]int)

	sem := make(chan struct{}, 5) // semaphore to limit concurrency
	var wg sync.WaitGroup         // WaitGroup to wait for all goroutines to finish

	for _, root := range roots {
		testFiles, err := os.ReadDir(root)
		require.NoError(t, err)

		for _, dirEntry := range testFiles {
			if dirEntry.IsDir() || isGeneratedFile(dirEntry.Name()) {
				continue
			}

			name := dirEntry.Name()
			path := filepath.Join(root, name)

			errorExpected := false
			if name == "range_check.small.cairo" {
				errorExpected = true
			}

			if !filter.filtered(name) {
				continue
			}

			// we run tests concurrently if we don't need benchmarks
			if !*zerobench {
				sem <- struct{}{} // acquire a semaphore slot
				wg.Add(1)

				go func(path, name string) {
					defer wg.Done()
					defer func() { <-sem }() // release the semaphore slot when done
					runAndTestFile(t, path, name, benchmarkMap, *zerobench, errorExpected)
				}(path, name)
			} else {
				runAndTestFile(t, path, name, benchmarkMap, *zerobench, errorExpected)
			}
		}
	}

	wg.Wait() // wait for all goroutines to finish

	for _, root := range roots {
		clean(root)
	}

	if *zerobench {
		WriteBenchMarksToFile(benchmarkMap)
	}
}

// Save the Benchmarks for the integration tests in `BenchMarks.txt`
func WriteBenchMarksToFile(benchmarkMap map[string][2]int) {
	totalWidth := 123

	border := strings.Repeat("=", totalWidth)
	separator := strings.Repeat("-", totalWidth)

	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 40, 0, 0, ' ', tabwriter.Debug)

	sb.WriteString(border + "\n")
	fmt.Fprintln(w, "| File \t PythonVM (ms) \t GoVM (ms) \t")
	w.Flush()
	sb.WriteString(border + "\n")

	iterator := 0
	totalFiles := len(benchmarkMap)

	for key, values := range benchmarkMap {
		row := "| " + key + "\t "

		for iter, value := range values {
			row = row + strconv.Itoa(value) + "\t"
			if iter == 0 {
				row = row + " "
			}
		}

		fmt.Fprintln(w, row)
		w.Flush()

		if iterator < totalFiles-1 {
			sb.WriteString(separator + "\n")
		}

		iterator++
	}

	sb.WriteString(border + "\n")

	fileName := "BenchMarks.txt"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file: ", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(sb.String())
	if err != nil {
		fmt.Println("Error writing to file: ", err)
	} else {
		fmt.Println("Benchmarks successfully written to:", fileName)
	}
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
func runPythonVm(testFilename, path string) (time.Duration, string, string, error) {
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
	} else if strings.HasSuffix(testFilename, ".dex.cairo") {
		args = append(args, "--layout", "dex")
	} else if strings.HasSuffix(testFilename, ".recursive.cairo") {
		args = append(args, "--layout", "recursive")
	} else if strings.HasSuffix(testFilename, ".starknet_with_keccak.cairo") {
		args = append(args, "--layout", "starknet_with_keccak")
	} else if strings.HasSuffix(testFilename, ".starknet.cairo") {
		args = append(args, "--layout", "starknet")
	} else if strings.HasSuffix(testFilename, ".recursive_large_output.cairo") {
		args = append(args, "--layout", "recursive_large_output")
	} else if strings.HasSuffix(testFilename, ".recursive_with_poseidon.cairo") {
		args = append(args, "--layout", "recursive_with_poseidon")
	} else if strings.HasSuffix(testFilename, ".all_solidity.cairo") {
		args = append(args, "--layout", "all_solidity")
	} else if strings.HasSuffix(testFilename, ".all_cairo.cairo") {
		args = append(args, "--layout", "all_cairo")
	}

	cmd := exec.Command("cairo-run", args...)

	start := time.Now()

	res, err := cmd.CombinedOutput()

	elapsed := time.Since(start)

	if err != nil {
		return 0, "", "", fmt.Errorf(
			"cairo-run %s: %w\n%s", path, err, string(res),
		)
	}

	return elapsed, traceOutput, memoryOutput, nil
}

// given a path to a compiled cairo zero file, execute
// it using our vm
func runVm(path string) (time.Duration, string, string, string, error) {
	traceOutput := swapExtenstion(path, traceSuffix)
	memoryOutput := swapExtenstion(path, memorySuffix)

	// If any other layouts are needed, add the suffix checks here.
	// The convention would be: ".$layout.cairo"
	// A file without this suffix will use the default ("plain") layout, which is a layout with no builtins included"
	layout := "plain"
	if strings.Contains(path, ".small") {
		layout = "small"
	} else if strings.Contains(path, ".dex") {
		layout = "dex"
	} else if strings.Contains(path, ".recursive") {
		layout = "recursive"
	} else if strings.Contains(path, ".starknet_with_keccak") {
		layout = "starknet_with_keccak"
	} else if strings.Contains(path, ".starknet") {
		layout = "starknet"
	} else if strings.Contains(path, ".recursive_large_output") {
		layout = "recursive_large_output"
	} else if strings.Contains(path, ".recursive_with_poseidon") {
		layout = "recursive_with_poseidon"
	} else if strings.Contains(path, ".all_solidity") {
		layout = "all_solidity"
	} else if strings.Contains(path, ".all_cairo") {
		layout = "all_cairo"
	}

	cmd := exec.Command(
		"../bin/cairo-vm",
		"run",
		"--proofmode",
		"--tracefile",
		traceOutput,
		"--memoryfile",
		memoryOutput,
		"--layout",
		layout,
		path,
	)

	start := time.Now()

	res, err := cmd.CombinedOutput()

	elapsed := time.Since(start)

	if err != nil {
		return 0, "", "", string(res), fmt.Errorf(
			"cairo-vm run %s: %w\n%s", path, err, string(res),
		)
	}

	return elapsed, traceOutput, memoryOutput, string(res), nil
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
