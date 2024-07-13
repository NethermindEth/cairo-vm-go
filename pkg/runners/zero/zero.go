package zero

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
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
	layout      builtins.Layout
}

// Creates a new Runner of a Cairo Zero program
func NewRunner(program *Program, hints map[uint64][]hinter.Hinter, proofmode bool, maxsteps uint64, layoutName string) (ZeroRunner, error) {
	hintrunner := hintrunner.NewHintRunner(hints)
	layout, err := builtins.GetLayout(layoutName)
	if err != nil {
		return ZeroRunner{}, err
	}
	return ZeroRunner{
		program:    program,
		hintrunner: hintrunner,
		proofmode:  proofmode,
		maxsteps:   maxsteps,
		layout:     layout,
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

		stack, err := runner.initializeBuiltins(memory)
		if err != nil {
			return mem.UnknownAddress, err
		}
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
	stack, err := runner.initializeBuiltins(memory)
	if err != nil {
		return mem.UnknownAddress, err
	}
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

func (runner *ZeroRunner) initializeBuiltins(memory *mem.Memory) ([]mem.MemoryValue, error) {
	builtinsSet := make(map[starknet.Builtin]bool)
	for _, bRunner := range runner.layout.Builtins {
		builtinsSet[bRunner.Builtin] = true
	}
	// check if all builtins from the program are in the layout
	for _, programBuiltin := range runner.program.Builtins {
		if _, found := builtinsSet[programBuiltin]; !found {
			builtinName, err := programBuiltin.MarshalJSON()
			if err != nil {
				return []mem.MemoryValue{}, err
			}
			return []mem.MemoryValue{}, fmt.Errorf("builtin %s not found in the layout: %s", builtinName, runner.layout.Name)
		}
	}
	stack := []mem.MemoryValue{}
	// adding to the stack only the builtins that are both in the program and in the layout
	for _, bRunner := range runner.layout.Builtins {
		builtinSegment := memory.AllocateBuiltinSegment(bRunner.Runner)
		if utils.Contains(runner.program.Builtins, bRunner.Builtin) {
			stack = append(stack, mem.MemoryValueFromMemoryAddress(&builtinSegment))
		}
	}
	return stack, nil
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

// EndRun is responsible for running the additional steps after the program was executed,
// until the checkUsedCells doesn't return any error.
// Since this vm always finishes the run of the program at the number of steps that is a power of two in the proof mode,
// there is no need to run additional steps before the loop.
func (runner *ZeroRunner) EndRun() error {
	if runner.proofmode {
		for runner.checkUsedCells() != nil {
			pow2Steps := utils.NextPowerOfTwo(runner.vm.Step + 1)
			if err := runner.RunFor(pow2Steps); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkUsedCells returns error if not enough steps were made to allocate required number of cells for builtins
// or there are not enough trace cells to fill the entire range check range
func (runner *ZeroRunner) checkUsedCells() error {
	for _, bRunner := range runner.layout.Builtins {
		builtinSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin(bRunner.Runner.String())
		if ok {
			_, err := bRunner.Runner.GetAllocatedSize(builtinSegment.Len(), runner.steps())
			if err != nil {
				return err
			}
		}
	}
	return runner.checkRangeCheckUsage()
}

// Checks if there are not enough trace cells to fill the entire range check range. Each step has assigned a number of range check units. If the number of unused range check units is less than the range of potential values to be checked (defined by rcMin and rcMax), the number of trace cells must be increased, by running additional steps.
func (runner *ZeroRunner) checkRangeCheckUsage() error {
	rcMin, rcMax := runner.getPermRangeCheckLimits()
	var rcUnitsUsedByBuiltins uint64
	for _, builtin := range runner.program.Builtins {
		if builtin == starknet.RangeCheck {
			for _, layoutBuiltin := range runner.layout.Builtins {
				if builtin == layoutBuiltin.Builtin {
					rangeCheckRunner, ok := layoutBuiltin.Runner.(*builtins.RangeCheck)
					if !ok {
						return fmt.Errorf("error type casting to *builtins.RangeCheck")
					}
					rangeCheckSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin(rangeCheckRunner.String())
					if ok {
						rcUnitsUsedByBuiltins += rangeCheckSegment.Len() * builtins.RANGE_CHECK_N_PARTS
					}
				}
			}
		}
	}
	// Out of the range check units allowed per step three are used for the instruction.
	unusedRcUnits := (runner.layout.RcUnits-3)*runner.vm.Step - rcUnitsUsedByBuiltins
	rcUsageUpperBound := uint64(rcMax - rcMin)
	if unusedRcUnits < rcUsageUpperBound {
		return fmt.Errorf("RangeCheck usage is %d, but the upper bound is %d", unusedRcUnits, rcUsageUpperBound)
	}
	return nil
}

// getPermRangeCheckLimits returns the minimum and maximum values used by the range check units in the program. To find the values, maximum and minimum values from the range check segment are compared with maximum and minimum values of instructions offsets calculated during running the instructions.
func (runner *ZeroRunner) getPermRangeCheckLimits() (uint16, uint16) {
	rcMin, rcMax := runner.vm.RcLimitsMin, runner.vm.RcLimitsMax

	for _, builtin := range runner.program.Builtins {
		if builtin == starknet.RangeCheck {
			bRunner := builtins.Runner(builtin)
			rangeCheckRunner, _ := bRunner.(*builtins.RangeCheck)
			rangeCheckSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin(rangeCheckRunner.String())
			if ok {
				rangeCheckUsageMin, rangeCheckUsageMax := rangeCheckRunner.GetRangeCheckUsage(rangeCheckSegment)
				if rangeCheckUsageMin < rcMin {
					rcMin = rangeCheckUsageMin
				}
				if rangeCheckUsageMax > rcMax {
					rcMax = rangeCheckUsageMax
				}
			}
		}
	}
	return rcMin, rcMax
}

// FinalizeSegments calculates the final size of the builtins segments,
// using number of allocated instances and memory cells per builtin instance.
// Additionally it sets the final size of the program segment to the program size.
func (runner *ZeroRunner) FinalizeSegments() error {
	programSize := uint64(len(runner.program.Bytecode))
	runner.vm.Memory.Segments[vm.ProgramSegment].Finalize(programSize)
	for _, bRunner := range runner.layout.Builtins {
		builtinSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin(bRunner.Runner.String())
		if ok {
			size, err := bRunner.Runner.GetAllocatedSize(builtinSegment.Len(), runner.vm.Step)
			if err != nil {
				return fmt.Errorf("builtin %s: %v", bRunner.Runner.String(), err)
			}
			builtinSegment.Finalize(size)
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
