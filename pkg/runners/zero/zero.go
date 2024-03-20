package zero

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type ZeroRunner struct {
	// core components
	program    *Program
	vm         *vm.VirtualMachine
	hintrunner hintrunner.HintRunner
	// config
	proofmode bool
	maxsteps  uint64
	// auxiliar
	runFinished bool
}

// Creates a new Runner of a Cairo Zero program
func NewRunner(program *Program, hints map[uint64][]hinter.Hinter, proofmode bool, maxsteps uint64) (ZeroRunner, error) {
	hintrunner := hintrunner.NewHintRunner(hints)

	return ZeroRunner{
		program:    program,
		hintrunner: hintrunner,
		proofmode:  proofmode,
		maxsteps:   maxsteps,
	}, nil
}

// RunEntryPoint is like Run, but it executes the program starting from the given PC offset.
// This PC offset is expected to be a start from some function inside the loaded program.
func (runner *ZeroRunner) RunEntryPoint(pc uint64) error {
	if runner.runFinished {
		return errors.New("cannot re-run using the same runner")
	}

	memory, err := runner.initializeSegments()
	if err != nil {
		return err
	}

	// Builtins are initialized as a part of initializeEntrypoint().

	returnFp := memory.AllocateEmptySegment()
	mvReturnFp := mem.MemoryValueFromMemoryAddress(&returnFp)
	end, err := runner.initializeEntrypoint(pc, nil, &mvReturnFp, memory)
	if err != nil {
		return err
	}

	if err := runner.RunUntilPc(&end); err != nil {
		return err
	}

	return nil
}

func (runner *ZeroRunner) Run() error {
	if runner.runFinished {
		return errors.New("cannot re-run using the same runner")
	}

	end, err := runner.InitializeMainEntrypoint()
	if err != nil {
		return fmt.Errorf("initializing main entry point: %w", err)
	}

	err = runner.RunUntilPc(&end)
	if err != nil {
		return err
	}

	if runner.proofmode {
		// +1 because proof mode require an extra instruction run
		// pow2 because proof mode also requires that the trace is a power of two
		pow2Steps := utils.NextPowerOfTwo(runner.vm.Step + 1)
		if err := runner.RunFor(pow2Steps); err != nil {
			return err
		}
	}
	return nil
}

func (runner *ZeroRunner) initializeSegments() (*mem.Memory, error) {
	memory := mem.InitializeEmptyMemory()
	_, err := memory.AllocateSegment(runner.program.Bytecode) // ProgramSegment
	if err != nil {
		return nil, err
	}

	memory.AllocateEmptySegment() // ExecutionSegment
	return memory, nil
}

// TODO: unexport it. It's only used inside this file and tests so far.
// We probably don't want various init API to leak outside (see #237 for more context).
func (runner *ZeroRunner) InitializeMainEntrypoint() (mem.MemoryAddress, error) {
	memory, err := runner.initializeSegments()
	if err != nil {
		return mem.UnknownAddress, err
	}

	if runner.proofmode {
		initialPCOffset, ok := runner.program.Labels["__start__"]
		if !ok {
			return mem.UnknownAddress,
				errors.New("start label not found. Try compiling with `--proof_mode`")
		}
		endPcOffset, ok := runner.program.Labels["__end__"]
		if !ok {
			return mem.UnknownAddress,
				errors.New("end label not found. Try compiling with `--proof_mode`")
		}

		stack := runner.initializeBuiltins(memory)
		// Add the dummy last fp and pc to the public memory, so that the verifier can enforce [fp - 2] = fp.
		stack = append([]mem.MemoryValue{mem.MemoryValueFromSegmentAndOffset(
			vm.ProgramSegment,
			len(runner.program.Bytecode)+2,
		), mem.EmptyMemoryValueAsFelt()}, stack...)

		if err := runner.initializeVm(&mem.MemoryAddress{
			SegmentIndex: vm.ProgramSegment,
			Offset:       initialPCOffset,
		}, stack, memory); err != nil {
			return mem.UnknownAddress, err
		}

		// __start__ will advance Ap and Fp
		runner.vm.Context.Ap = 2
		runner.vm.Context.Fp = 2
		return mem.MemoryAddress{SegmentIndex: vm.ProgramSegment, Offset: endPcOffset}, nil
	}

	returnFp := memory.AllocateEmptySegment()
	mvReturnFp := mem.MemoryValueFromMemoryAddress(&returnFp)
	mainPCOffset, ok := runner.program.Entrypoints["main"]
	if !ok {
		return mem.UnknownAddress, errors.New("can't find an entrypoint for main")
	}
	return runner.initializeEntrypoint(mainPCOffset, nil, &mvReturnFp, memory)
}

func (runner *ZeroRunner) initializeEntrypoint(
	initialPCOffset uint64, arguments []*f.Element, returnFp *mem.MemoryValue, memory *mem.Memory,
) (mem.MemoryAddress, error) {
	stack := runner.initializeBuiltins(memory)
	for i := range arguments {
		stack = append(stack, mem.MemoryValueFromFieldElement(arguments[i]))
	}
	end := memory.AllocateEmptySegment()

	stack = append(stack, *returnFp, mem.MemoryValueFromMemoryAddress(&end))
	return end, runner.initializeVm(&mem.MemoryAddress{
		SegmentIndex: vm.ProgramSegment,
		Offset:       initialPCOffset,
	}, stack, memory)
}

func (runner *ZeroRunner) initializeBuiltins(memory *mem.Memory) []mem.MemoryValue {
	stack := []mem.MemoryValue{}
	for _, builtin := range runner.program.Builtins {
		bRunner := builtins.Runner(builtin)
		builtinSegment := memory.AllocateBuiltinSegment(bRunner)
		stack = append(stack, mem.MemoryValueFromMemoryAddress(&builtinSegment))
	}
	return stack
}

func (runner *ZeroRunner) initializeVm(
	initialPC *mem.MemoryAddress, stack []mem.MemoryValue, memory *mem.Memory,
) error {
	executionSegment := memory.Segments[vm.ExecutionSegment]
	offset := executionSegment.Len()
	for idx := range stack {
		if err := executionSegment.Write(offset+uint64(idx), &stack[idx]); err != nil {
			return err
		}
	}

	var err error
	// initialize vm
	runner.vm, err = vm.NewVirtualMachine(vm.Context{
		Pc: *initialPC,
		Ap: offset + uint64(len(stack)),
		Fp: offset + uint64(len(stack)),
	}, memory, vm.VirtualMachineConfig{ProofMode: runner.proofmode})
	return err
}

// run until the program counter equals the `pc` parameter
func (runner *ZeroRunner) RunUntilPc(pc *mem.MemoryAddress) error {
	for !runner.vm.Context.Pc.Equal(pc) {
		if runner.steps() >= runner.maxsteps {
			return fmt.Errorf(
				"pc %s step %d: max step limit exceeded (%d)",
				runner.pc(),
				runner.steps(),
				runner.maxsteps,
			)
		}
		if err := runner.vm.RunStep(&runner.hintrunner); err != nil {
			return fmt.Errorf("pc %s step %d: %w", runner.pc(), runner.steps(), err)
		}
	}
	return nil
}

// run until the vm step count reaches the `steps` parameter
func (runner *ZeroRunner) RunFor(steps uint64) error {
	for runner.steps() < steps {
		if runner.steps() >= runner.maxsteps {
			return fmt.Errorf(
				"pc %s step %d: max step limit exceeded (%d)",
				runner.pc(),
				runner.steps(),
				runner.maxsteps,
			)
		}
		if err := runner.vm.RunStep(&runner.hintrunner); err != nil {
			return fmt.Errorf(
				"pc %s step %d: %w",
				runner.pc(),
				runner.steps(),
				err,
			)
		}
	}
	return nil
}

func (runner *ZeroRunner) BuildProof() ([]byte, []byte, error) {
	relocatedTrace, err := runner.vm.ExecutionTrace()
	if err != nil {
		return nil, nil, err
	}

	return vm.EncodeTrace(relocatedTrace), vm.EncodeMemory(runner.vm.RelocateMemory()), nil
}

func (runner *ZeroRunner) pc() mem.MemoryAddress {
	return runner.vm.Context.Pc
}

func (runner *ZeroRunner) steps() uint64 {
	return runner.vm.Step
}

// Gives the output of the last run. Panics if there hasn't
// been any runs yet.
func (runner *ZeroRunner) Output() []*fp.Element {
	if runner.vm == nil {
		panic("cannot get the output from an uninitialized runner")
	}

	output := []*fp.Element{}
	outputSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin("output")
	if !ok {
		return output
	}

	for offset := uint64(0); offset < outputSegment.Len(); offset++ {
		value := outputSegment.Peek(offset)
		// no need to check for an error here since only felts can be written
		// to the output segment
		valueFelt, _ := value.FieldElement()
		output = append(output, valueFelt)
	}
	return output
}
