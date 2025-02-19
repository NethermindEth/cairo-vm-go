package integrationtests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/joho/godotenv"
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

func WriteBenchMarksToFile(benchmarkMap map[string][]int) {
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
	compiledSuffix       = "_compiled.json"
	sierraSuffix         = ".sierra"
	pyTraceSuffix        = "_py_trace"
	pyMemorySuffix       = "_py_memory"
	rsTraceSuffix        = "_rs_trace"
	rsMemorySuffix       = "_rs_memory"
	traceSuffix          = "_trace"
	memorySuffix         = "_memory"
	airPublicInputSuffix = "_air_public_input.json"
)

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

func isGeneratedFile(path string) bool {
	return strings.HasSuffix(path, compiledSuffix) ||
		strings.HasSuffix(path, sierraSuffix) ||
		strings.HasSuffix(path, pyTraceSuffix) ||
		strings.HasSuffix(path, pyMemorySuffix) ||
		strings.HasSuffix(path, traceSuffix) ||
		strings.HasSuffix(path, memorySuffix) ||
		strings.HasSuffix(path, airPublicInputSuffix)
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

// Given a certain file, it swaps its extension with a new suffix
func swapExtenstion(path string, newSuffix string) string {
	dir, name := filepath.Split(path)
	name = name[:len(name)-len(".cairo")]
	name = fmt.Sprintf("%s%s", name, newSuffix)
	return filepath.Join(dir, name)
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
