package integrationtests

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runAndTestCairoZeroFile(t *testing.T, path string, name string, benchmarkMap map[string][]int, benchmark bool, errorExpected bool) {
	t.Logf("testing: %s\n", path)
	compiledOutput, err := compileCairoZeroCode(path)
	if err != nil {
		t.Error(err)
		return
	}
	layout := getLayoutFromFileName(path)

	elapsedGo, traceFile, memoryFile, _, err := runVmZero(compiledOutput, layout)
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

	rustVmFilePath := compiledOutput
	elapsedRs, rsTraceFile, rsMemoryFile, err := runRustVmZero(rustVmFilePath, layout)
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

	elapsedPy, pyTraceFile, pyMemoryFile, err := runPythonVmZero(compiledOutput, layout)
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
		benchmarkMap[name] = []int{int(elapsedPy.Milliseconds()), int(elapsedGo.Milliseconds()), int(elapsedRs.Milliseconds())}
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

var zerobench = flag.Bool("zerobench", false, "run integration tests and generate benchmarks file")

func TestCairoZeroFiles(t *testing.T) {
	file, err := os.OpenFile(whitelistFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open file: %w", err))
	}

	file.Close()

	roots := []string{
		"./cairo_zero_hint_tests/",
		"./cairo_zero_file_tests/",
		"./builtin_tests/",
	}

	// filter is for debugging purposes
	filter := Filter{}
	filter.init()

	benchmarkMap := make(map[string][]int)

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
			if name == "range_check__small.cairo" {
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
					runAndTestCairoZeroFile(t, path, name, benchmarkMap, *zerobench, errorExpected)
				}(path, name)
			} else {
				runAndTestCairoZeroFile(t, path, name, benchmarkMap, *zerobench, errorExpected)
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

// given a path to a cairo zero file, it compiles it
// and return the compilation path
func compileCairoZeroCode(path string) (string, error) {
	if filepath.Ext(path) != ".cairo" {
		return "", fmt.Errorf("compiling non cairo file: %s", path)
	}
	compiledOutput := swapExtenstion(path, compiledSuffix)

	var cliCommand string
	var args []string

	cliCommand = "cairo-compile"
	args = []string{
		path,
		"--proof_mode",
		"--no_debug_info",
		"--output",
		compiledOutput,
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
func runPythonVmZero(path, layout string) (time.Duration, string, string, error) {
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
func runRustVmZero(path, layout string) (time.Duration, string, string, error) {
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
		"--proof_mode",
	}

	binaryPath := "./../rust_vm_bin/cairo-vm-cli"

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
func runVmZero(path, layout string) (time.Duration, string, string, string, error) {
	traceOutput := swapExtenstion(path, traceSuffix)
	memoryOutput := swapExtenstion(path, memorySuffix)

	args := []string{
		"run",
		"--proofmode",
		"--tracefile",
		traceOutput,
		"--memoryfile",
		memoryOutput,
		"--layout",
		layout,
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
