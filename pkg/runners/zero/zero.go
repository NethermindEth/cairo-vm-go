package runner

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Program struct {
	// the bytecode in string format
	Bytecode    []*f.Element
	Identifiers []any
	// given a string it returns the pc for that function call
	Entrypoints map[string]uint64
}

type ZeroRunner struct {
	program    *Program
	vm         *VM.VirtualMachine
	hintrunner VM.HintRunner
	proofmode  bool
}

func LoadCairoZeroProgram(content []byte) *Program {
	// parse the cairo zero file
	// return Load with the Program to parse
	return nil
}

// Creates a new Runner of a Cairo Zero program
func NewRunner(program *Program, proofmode bool) (*ZeroRunner, error) {
	//var bytecode []*f.Element
	//for i := range program.Bytecode {
	//	felt, err := new(f.Element).SetString(program.Bytecode[i])
	//	if err != nil {
	//		return fmt.Errorf(
	//			"runner error: cannot read bytecode %s at position %d: %w",
	//			program.Bytecode[i], i, err,
	//		)
	//	}
	//	bytecode[i] = felt
	//}

	// initialize vm
	vm, err := VM.NewVirtualMachine(program.Bytecode, VM.VirtualMachineConfig{ProofMode: proofmode})
	if err != nil {
		return nil, fmt.Errorf("runner error: %w", err)
	}

	// intialize hintrunner
	// todo(rodro): given the program get the appropiate hints
	hintrunner := hintrunner.NewHintRunner(make(map[uint64]hintrunner.Hinter))

	return &ZeroRunner{
		program:    program,
		vm:         vm,
		hintrunner: hintrunner,
		proofmode:  proofmode,
	}, nil
}

func (runner *ZeroRunner) InitializeMainEntrypoint() (uint64, error) {
	if runner.proofmode {
		panic("runner proofmode not implemented yet")
	} else {
		returnFp := memory.MemoryValueFromSegmentAndOffset(
			runner.memory().AllocateEmptySegment(),
			0,
		)
		return runner.InitializeEntrypoint("main", nil, returnFp)
	}
}

func (runner *ZeroRunner) InitializeEntrypoint(
	funcName string, arguments []*f.Element, returnFp *memory.MemoryValue,
) (uint64, error) {
	end := runner.segments()[VM.ProgramSegment].Len()
	// write arguments
	for i := range arguments {
		err := runner.memory().Write(VM.ExecutionSegment, uint64(i), memory.MemoryValueFromFieldElement(arguments[i]))
		if err != nil {
			return 0, err
		}
	}
	offset := runner.segments()[VM.ExecutionSegment].Len()
	err := runner.memory().Write(VM.ExecutionSegment, offset, returnFp)
	if err != nil {
		return 0, err
	}
	err = runner.memory().Write(VM.ExecutionSegment, offset+1, memory.MemoryValueFromInt(end))
	if err != nil {
		return 0, err
	}

	pc, ok := runner.program.Entrypoints[funcName]
	if !ok {
		return 0, fmt.Errorf("unknwon entrypoint: %s", funcName)
	}

	runner.vm.Context.Pc = pc
	runner.vm.Context.Ap = offset + 2
	runner.vm.Context.Fp = runner.vm.Context.Ap

	return end, nil
}

func (runner *ZeroRunner) RunUntilPc(pc uint64) error {
	for runner.vm.Context.Pc != pc {
		err := runner.vm.RunStep(runner.hintrunner)
		if err != nil {
			return err
		}
	}
	return nil
}

func (runner *ZeroRunner) BuildProof() error {
	panic("not implemented yet")
	//_, _, err := runner.vm.Proof()
	//if err != nil {
	//	return err
	//}
	//return nil
}

func (runner *ZeroRunner) memory() *memory.Memory {
	return runner.vm.MemoryManager.Memory
}

func (runner *ZeroRunner) segments() []*memory.Segment {
	return runner.vm.MemoryManager.Memory.Segments
}
