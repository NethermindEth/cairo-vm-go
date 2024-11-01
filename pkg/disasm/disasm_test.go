package disasm_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	"github.com/NethermindEth/cairo-vm-go/pkg/disasm"
	"github.com/stretchr/testify/assert"
)

func TestDisasm(t *testing.T) {
	tests, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	// This code tests both CASM parser and disassembler.
	// * We take some pre-existing CASM file from the test suite
	// * Then it's parsed by our assembler package (instList)
	// * The result of our parser is used as a disasm input
	// * The disassembly is then stringified to get a new CASM input file
	// * This new CASM file is given to our assembler package (instList2)
	// * As a final step, instList and instList2 slices are compared
	//
	// If they're equal, it's a good sign and we have a high chance
	// of asm+disasm not losing important information.
	//
	// We could also check for testCode and testCode2 being identical,
	// but apart from the formatting issues (easy to solve) there are
	// equivalent ways to write some expressions, e.g.:
	// > [fp + -3] vs [fp - 3]
	// > [ap+0] vs [ap]
	// Both of them encode the same dereference expr, just spelled differently.
	// But anyway, with an extra flag we can mark some of the tests as
	// "ok to check for textual equallity" (see checkTexts map below).

	checkTexts := map[string]struct{}{
		"simple.casm":            {},
		"hash_chain_pretty.casm": {},
	}

	disasmProgToString := func(p *disasm.Program) string {
		// Ignore the comments.
		var buf strings.Builder
		for _, l := range p.Lines {
			if len(l.Text) == 0 {
				// This is a comment-only line.
				continue
			}
			fmt.Fprintf(&buf, "%s;\n", l.Text)
		}
		return buf.String()
	}

	for _, test := range tests {
		testName := filepath.Base(test.Name())
		testFile := filepath.Join("testdata", test.Name())
		t.Run(testName, func(t *testing.T) {
			testCode, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatal(err)
			}
			instList, _, err := assembler.CasmToBytecode(string(testCode))
			if err != nil {
				t.Fatal(err)
			}
			disassembled, err := disasm.FromBytecode(disasm.Config{
				Bytecode: instList,
				Indent:   0,
			})
			if err != nil {
				t.Fatal(err)
			}
			testCode2 := disasmProgToString(disassembled)
			instList2, _, err := assembler.CasmToBytecode(testCode2)
			if err != nil {
				t.Fatalf("generated casm parse error: %v\nprog:\n%s", err, testCode2)
			}
			assert.Equal(t, instList, instList2)
			if _, ok := checkTexts[testName]; ok {
				assert.Equal(t, string(testCode), testCode2)
			}
		})
	}
}
