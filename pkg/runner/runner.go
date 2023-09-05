package runner

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"go/parser"
	"text/scanner"
)

type Program struct {
	// the bytecode in string format
	Bytecode []string
	// given a string it returns the pc for that function call
	Entrypoints map[string]uint64
}

type Runner struct {
	program Program
}

func LoadCairoZeroProgram(content []byte) *Program {
	// parse the cairo zero file
	// return Load with the Program to parse
	return nil
}

// It recieves the program to run, the function name of this program and the arguments passed
// as inputs
func Run(program Program, entrypoint string, arguments []string) error {
	var bytecode []*f.Element
	for i := range program.Bytecode {
		felt, err := new(f.Element).SetString(program.Bytecode[i])
		if err != nil {
			return NewRunnerError(err)
		}
		bytecode[i] = felt
	}

	// initialize vm
	vm, err := VM.NewVirtualMachine(bytecode, VM.VirtualMachineConfig{})
	if err != nil {
		return NewRunnerError(err)
	}

	// intialize hintrunner
	hintrunner := hintrunner.NewHintRunner(make(map[uint64]hintrunner.Hinter))

	// set the appropriate pc using the Entrypoints info
	vm.Context.Pc = program.Entrypoints[entrypoint]

	// We need to store all arguments as felts before starting execution
	// if there are n felts used as arguments, then fp and ap should start at n
	// and [0, n - 1] should have all argument information in order
	for i := range arguments {
		felt, err := new(f.Element).SetString(arguments[i])
		if err != nil {
			return NewRunnerError(err)
		}
		vm.MemoryManager.Memory.Write(VM.ExecutionSegment, uint64(i), memory.MemoryValueFromFieldElement(felt))
	}
	vm.Context.Ap = uint64(len(arguments))
	vm.Context.Fp = vm.Context.Ap

	return nil
}
