package integrationtests

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

const whitelistFile = "./list_tests_in_progress.txt"

func writeToFile(content string) {
	file, err := os.OpenFile(whitelistFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open file: %w", err))
	}
	defer file.Close()

	if _, err := file.WriteString(content + "\n"); err != nil {
		panic(fmt.Errorf("failed to write to file: %w", err))
	}
}

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

func runAndTestFile(t *testing.T, path string, name string, benchmarkMap map[string][3]int, benchmark bool, errorExpected bool, zero bool, inputArgs string) {
	t.Logf("testing: %s\n", path)
	compiledOutput, err := compileCairoCode(path, zero)
	if err != nil {
		t.Error(err)
		return
	}
	layout := getLayoutFromFileName(path)

	elapsedGo, traceFile, memoryFile, _, err := runVm(compiledOutput, layout, zero, inputArgs)
	if errorExpected {
		assert.Error(t, err, path)
		writeToFile(path)
		return
	} else {
		if err != nil {
			t.Error(err)
			writeToFile(path)
			return
		}
	}

	rustVmFilePath := path
	if zero {
		rustVmFilePath = compiledOutput
	}
	elapsedRs, rsTraceFile, rsMemoryFile, err := runRustVm(rustVmFilePath, layout, zero, inputArgs)
	if errorExpected {
		// we let the code go on so that we can check if the go vm also raises an error
		assert.Error(t, err, path)
		return
	} else {
		if err != nil {
			t.Error(err)
			writeToFile(path)
			return
		}
	}

	trace, memory, err := decodeProof(traceFile, memoryFile)
	if err != nil {
		t.Error(err)
		writeToFile(path)
		return
	}
	rsTrace, rsMemory, err := decodeProof(rsTraceFile, rsMemoryFile)
	if err != nil {
		t.Error(err)
		writeToFile(path)
		return
	}

	if !assert.Equal(t, rsTrace, trace) {
		t.Logf("rstrace:\n%s\n", traceRepr(rsTrace))
		t.Logf("trace:\n%s\n", traceRepr(trace))
		writeToFile(path)
	}
	if !assert.Equal(t, rsMemory, memory) {
		t.Logf("rsmemory;\n%s\n", memoryRepr(rsMemory))
		t.Logf("memory;\n%s\n", memoryRepr(memory))
		writeToFile(path)
	}

	if zero {
		elapsedPy, pyTraceFile, pyMemoryFile, err := runPythonVm(compiledOutput, layout)
		if errorExpected {
			// we let the code go on so that we can check if the go vm also raises an error
			assert.Error(t, err, path)
		} else {
			if err != nil {
				t.Error(err)
				return
			}
		}

		if benchmark {
			benchmarkMap[name] = [3]int{int(elapsedPy.Milliseconds()), int(elapsedGo.Milliseconds()), int(elapsedRs.Milliseconds())}
		}

		pyTrace, pyMemory, err := decodeProof(pyTraceFile, pyMemoryFile)
		if err != nil {
			t.Error(err)
			return
		}

		if !assert.Equal(t, pyTrace, trace) {
			t.Logf("pytrace:\n%s\n", traceRepr(pyTrace))
			t.Logf("trace:\n%s\n", traceRepr(trace))
			writeToFile(path)
		}
		if !assert.Equal(t, pyMemory, memory) {
			t.Logf("pymemory;\n%s\n", memoryRepr(pyMemory))
			t.Logf("memory;\n%s\n", memoryRepr(memory))
			writeToFile(path)
		}
	}
}

var zerobench = flag.Bool("zerobench", false, "run integration tests and generate benchmarks file")

func TestCairoFiles(t *testing.T) {
	file, err := os.OpenFile(whitelistFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open file: %w", err))
	}

	file.Close()
	type TestCase struct {
		path string
		zero bool
	}
	roots := []TestCase{
		// {"./cairo_zero_hint_tests/", true},
		// {"./cairo_zero_file_tests/", true},
		// {"./builtin_tests/", true},
		{"./cairo_1_programs/", false},
		{"./cairo_1_programs/dict_non_squashed", false},
		{"./cairo_1_programs/with_input", false},
	}

	inputArgsMap := map[string]string{
		"cairo_1_programs/with_input/array_input_sum__small.cairo": "2 [111 222 333] 1 [444 555 666 777]",
		"cairo_1_programs/with_input/array_length__small.cairo":    "[1 2 3 4 5 6] [7 8 9 10]",
		"cairo_1_programs/with_input/branching.cairo":              "123",
		"cairo_1_programs/with_input/dict_with_input__small.cairo": "[1 2 3 4]",
		"cairo_1_programs/with_input/tensor__small.cairo":          "[1 4] [1 5]",
	}

	// filter is for debugging purposes
	filter := Filter{}
	filter.init()

	benchmarkMap := make(map[string][3]int)

	sem := make(chan struct{}, 5) // semaphore to limit concurrency
	var wg sync.WaitGroup         // WaitGroup to wait for all goroutines to finish

	for _, root := range roots {
		testFiles, err := os.ReadDir(root.path)
		require.NoError(t, err)

		for _, dirEntry := range testFiles {
			if dirEntry.IsDir() || isGeneratedFile(dirEntry.Name()) {
				continue
			}

			name := dirEntry.Name()
			path := filepath.Join(root.path, name)

			errorExpected := false
			if name == "range_check__small.cairo" {
				errorExpected = true
			}
			if !filter.filtered(name) {
				continue
			}
			inputArgs := inputArgsMap[path]
			// we run tests concurrently if we don't need benchmarks
			if !*zerobench {
				sem <- struct{}{} // acquire a semaphore slot
				wg.Add(1)

				go func(path, name string, root TestCase, inputArgs string) {
					defer wg.Done()
					defer func() { <-sem }() // release the semaphore slot when done
					runAndTestFile(t, path, name, benchmarkMap, *zerobench, errorExpected, root.zero, inputArgs)
				}(path, name, root, inputArgs)
			} else {
				runAndTestFile(t, path, name, benchmarkMap, *zerobench, errorExpected, root.zero, inputArgs)
			}
		}
	}

	wg.Wait() // wait for all goroutines to finish

	for _, root := range roots {
		clean(root.path)
	}

	if *zerobench {
		WriteBenchMarksToFile(benchmarkMap)
	}
}

func WriteBenchMarksToFile(benchmarkMap map[string][3]int) {
	totalWidth := 113 // Reduced width to adjust for long file names

	border := strings.Repeat("=", totalWidth)
	separator := strings.Repeat("-", totalWidth)

	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', tabwriter.AlignRight)

	sb.WriteString(border + "\n")
	fmt.Fprintf(w, "| %-40s | %-20s | %-20s | %-20s |\n", "File", "PythonVM (ms)", "GoVM (ms)", "RustVM (ms)")
	w.Flush()
	sb.WriteString(border + "\n")

	iterator := 0
	totalFiles := len(benchmarkMap)

	for key, values := range benchmarkMap {
		// Adjust the key length if it's too long
		displayKey := key
		if len(displayKey) > 40 {
			displayKey = displayKey[:37] + "..."
		}

		fmt.Fprintf(w, "| %-40s | %-20d | %-20d | %-20d |\n", displayKey, values[0], values[1], values[2])
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
	sierraSuffix   = ".sierra"
	pyTraceSuffix  = "_py_trace"
	pyMemorySuffix = "_py_memory"
	rsTraceSuffix  = "_rs_trace"
	rsMemorySuffix = "_rs_memory"
	traceSuffix    = "_trace"
	memorySuffix   = "_memory"
)

// given a path to a cairo zero file, it compiles it
// and return the compilation path
func compileCairoCode(path string, zero bool) (string, error) {
	if filepath.Ext(path) != ".cairo" {
		return "", fmt.Errorf("compiling non cairo file: %s", path)
	}
	compiledOutput := swapExtenstion(path, compiledSuffix)

	var cliCommand string
	var args []string

	if zero {
		cliCommand = "cairo-compile"
		args = []string{
			path,
			"--proof_mode",
			"--no_debug_info",
			"--output",
			compiledOutput,
		}
	} else {
		sierraOutput := swapExtenstion(path, sierraSuffix)
		cliCommand = "../rust_vm_bin/cairo-lang/cairo-compile"
		args = []string{
			"--single-file",
			path,
			sierraOutput,
			"--replace-ids",
		}

		cmd := exec.Command(cliCommand, args...)

		res, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf(
				"%s %s: %w\n%s", cliCommand, path, err, string(res),
			)
		}

		cliCommand = "../rust_vm_bin/cairo-lang/sierra-compile-json"
		args = []string{
			sierraOutput,
			compiledOutput,
		}

	}
	cmd := exec.Command(cliCommand, args...)

	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"%s %s: %w\n%s", cliCommand, path, err, string(res),
		)
	}

	return compiledOutput, nil
}

// given a path to a compiled cairo zero file, execute it using the
// python vm and returns the trace and memory files location
func runPythonVm(path, layout string) (time.Duration, string, string, error) {
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
		"--layout",
		layout,
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

// given a path to a compiled cairo zero file, execute it using the
// rust vm and return the trace and memory files location
func runRustVm(path, layout string, zero bool, inputArgs string) (time.Duration, string, string, error) {
	traceOutput := swapExtenstion(path, rsTraceSuffix)
	memoryOutput := swapExtenstion(path, rsMemorySuffix)

	args := []string{
		path,
		"--trace_file",
		traceOutput,
		"--memory_file",
		memoryOutput,
		"--layout",
		layout,
		"--args",
		inputArgs,
	}

	if zero {
		args = append(args, "--proof_mode")
	}

	binaryPath := "./../rust_vm_bin/cairo-lang/cairo1-run"
	if zero {
		binaryPath = "./../rust_vm_bin/cairo-vm-cli"
	}
	cmd := exec.Command(binaryPath, args...)

	start := time.Now()

	res, err := cmd.CombinedOutput()

	elapsed := time.Since(start)

	if err != nil {
		return 0, "", "", fmt.Errorf(
			"%s %s: %w\n%s", binaryPath, path, err, string(res),
		)
	}

	return elapsed, traceOutput, memoryOutput, nil
}

// given a path to a compiled cairo zero file, execute
// it using our vm
func runVm(path, layout string, zero bool, inputArgs string) (time.Duration, string, string, string, error) {
	traceOutput := swapExtenstion(path, traceSuffix)
	memoryOutput := swapExtenstion(path, memorySuffix)

	cliCommand := "cairo-run"
	if zero {
		cliCommand = "run"
	}
	args := []string{
		cliCommand,
		"--proofmode",
		"--tracefile",
		traceOutput,
		"--memoryfile",
		memoryOutput,
		"--layout",
		layout,
	}
	if !zero {
		args = []string{
			cliCommand,
			// "--proofmode",
			"--tracefile",
			traceOutput,
			"--memoryfile",
			memoryOutput,
			"--layout",
			layout,
			"--available_gas",
			"9999999",
			"--args",
			inputArgs,
		}
	}
	args = append(args, path)
	cmd := exec.Command(
		"../bin/cairo-vm",
		args...,
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
		strings.HasSuffix(path, sierraSuffix) ||
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

// If any other layouts are needed, add the suffix checks here.
// The convention would be: ".$layout.cairo"
// A file without this suffix will use the default ("plain") layout, which is a layout with no builtins included"
func getLayoutFromFileName(path string) string {
	if strings.HasSuffix(path, "__small.cairo") {
		return "small"
	} else if strings.HasSuffix(path, "__dex.cairo") {
		return "dex"
	} else if strings.HasSuffix(path, "__recursive.cairo") {
		return "recursive"
	} else if strings.HasSuffix(path, "__starknet_with_keccak.cairo") {
		return "starknet_with_keccak"
	} else if strings.HasSuffix(path, "__starknet.cairo") {
		return "starknet"
	} else if strings.HasSuffix(path, "__recursive_large_output.cairo") {
		return "recursive_large_output"
	} else if strings.HasSuffix(path, "__recursive_with_poseidon.cairo") {
		return "recursive_with_poseidon"
	} else if strings.HasSuffix(path, "__all_solidity.cairo") {
		return "all_solidity"
	} else if strings.HasSuffix(path, "__all_cairo.cairo") {
		return "all_cairo"
	}
	return "plain"
}
