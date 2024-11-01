package runner

import (
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type RunnerMode uint8

const (
	ExecutionMode RunnerMode = iota + 1
	ProofModeCairo0
	ProofModeCairo1
)

type Runner struct {
	// core components
	program    *Program
	vm         *vm.VirtualMachine
	hintrunner hintrunner.HintRunner
	// config
	collectTrace bool
	maxsteps     uint64
	runnerMode   RunnerMode
	// auxiliar
	runFinished bool
	layout      builtins.Layout
}

type CairoRunner struct{}

// Creates a new Runner of a Cairo Zero program
func NewRunner(program *Program, hints map[uint64][]hinter.Hinter, runnerMode RunnerMode, collectTrace bool, maxsteps uint64, layoutName string) (Runner, error) {
	hintrunner := hintrunner.NewHintRunner(hints)
	layout, err := builtins.GetLayout(layoutName)
	if err != nil {
		return Runner{}, err
	}
	return Runner{
		program:      program,
		runnerMode:   runnerMode,
		hintrunner:   hintrunner,
		collectTrace: collectTrace,
		maxsteps:     maxsteps,
		layout:       layout,
	}, nil
}

// RunEntryPoint is like Run, but it executes the program starting from the given PC offset.
// This PC offset is expected to be a start from some function inside the loaded program.
func (runner *Runner) RunEntryPoint(pc uint64) error {
	if runner.runFinished {
		return errors.New("cannot re-run using the same runner")
	}

	memory, err := runner.initializeSegments()
	if err != nil {
		return err
	}

	stack, err := runner.initializeBuiltins(memory)
	if err != nil {
		return err
	}

	// Builtins are initialized as a part of initializeEntrypoint().

	returnFp := memory.AllocateEmptySegment()
	mvReturnFp := mem.MemoryValueFromMemoryAddress(&returnFp)
	cairo1FpOffset := uint64(0)
	if runner.runnerMode == ProofModeCairo1 {
		cairo1FpOffset = 2
	}
	end, err := runner.initializeEntrypoint(pc, nil, &mvReturnFp, memory, stack, cairo1FpOffset)
	if err != nil {
		return err
	}

	if err := runner.RunUntilPc(&end); err != nil {
		return err
	}

	return nil
}

func (runner *Runner) Run() error {
	if runner.runFinished {
		return errors.New("cannot re-run using the same runner")
	}

	end, err := runner.initializeMainEntrypoint()
	if err != nil {
		return fmt.Errorf("initializing main entry point: %w", err)
	}

	err = runner.RunUntilPc(&end)
	if err != nil {
		return err
	}

	if runner.runnerMode == ProofModeCairo0 || runner.runnerMode == ProofModeCairo1 {
		// +1 because proof mode require an extra instruction run
		// pow2 because proof mode also requires that the trace is a power of two
		pow2Steps := utils.NextPowerOfTwo(runner.vm.Step + 1)
		if err := runner.RunFor(pow2Steps); err != nil {
			return err
		}
	}
	return nil
}

func (runner *Runner) initializeSegments() (*mem.Memory, error) {
	memory := mem.InitializeEmptyMemory()
	_, err := memory.AllocateSegment(runner.program.Bytecode) // ProgramSegment
	if err != nil {
		return nil, err
	}

	memory.AllocateEmptySegment() // ExecutionSegment
	return memory, nil
}

func (runner *Runner) initializeMainEntrypoint() (mem.MemoryAddress, error) {
	memory, err := runner.initializeSegments()
	if err != nil {
		return mem.UnknownAddress, err
	}

	stack, err := runner.initializeBuiltins(memory)
	if err != nil {
		return mem.UnknownAddress, err
	}
	switch runner.runnerMode {
	case ExecutionMode:
		returnFp := memory.AllocateEmptySegment()
		mvReturnFp := mem.MemoryValueFromMemoryAddress(&returnFp)
		mainPCOffset, ok := runner.program.Entrypoints["main"]
		if !ok {
			return mem.UnknownAddress, errors.New("can't find an entrypoint for main")
		}
		return runner.initializeEntrypoint(mainPCOffset, nil, &mvReturnFp, memory, stack, 0)
	case ProofModeCairo1:
		returnFp := memory.AllocateEmptySegment()
		mvReturnFp := mem.MemoryValueFromMemoryAddress(&returnFp)
		mainPCOffset, ok := runner.program.Entrypoints["main"]
		if !ok {
			return mem.UnknownAddress, errors.New("can't find an entrypoint for main")
		}
		return runner.initializeEntrypoint(mainPCOffset, nil, &mvReturnFp, memory, stack, 2)
	case ProofModeCairo0:
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

		// Add the dummy last fp and pc to the public memory, so that the verifier can enforce [fp - 2] = fp.
		stack = append([]mem.MemoryValue{mem.MemoryValueFromSegmentAndOffset(
			vm.ProgramSegment,
			len(runner.program.Bytecode)+2,
		), mem.EmptyMemoryValueAsFelt()}, stack...)

		if err := runner.initializeVm(&mem.MemoryAddress{
			SegmentIndex: vm.ProgramSegment,
			Offset:       initialPCOffset,
		}, stack, memory, 0); err != nil {
			return mem.UnknownAddress, err
		}

		// __start__ will advance Ap and Fp
		runner.vm.Context.Ap = 2
		runner.vm.Context.Fp = 2
		return mem.MemoryAddress{SegmentIndex: vm.ProgramSegment, Offset: endPcOffset}, nil

	}
	return mem.UnknownAddress, errors.New("unknown runner mode")
}

func (runner *Runner) initializeEntrypoint(
	initialPCOffset uint64, arguments []*fp.Element, returnFp *mem.MemoryValue, memory *mem.Memory, stack []mem.MemoryValue, cairo1FpOffset uint64,
) (mem.MemoryAddress, error) {
	for i := range arguments {
		stack = append(stack, mem.MemoryValueFromFieldElement(arguments[i]))
	}
	endPC := memory.AllocateEmptySegment()
	stack = append(stack, *returnFp, mem.MemoryValueFromMemoryAddress(&endPC))
	return endPC, runner.initializeVm(&mem.MemoryAddress{
		SegmentIndex: vm.ProgramSegment,
		Offset:       initialPCOffset,
	}, stack, memory, cairo1FpOffset)
}

func (runner *Runner) initializeBuiltins(memory *mem.Memory) ([]mem.MemoryValue, error) {
	builtinsSet := make(map[builtins.BuiltinType]bool)
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

func (runner *Runner) initializeVm(
	initialPC *mem.MemoryAddress, stack []mem.MemoryValue, memory *mem.Memory, cairo1FpOffset uint64,
) error {
	executionSegment := memory.Segments[vm.ExecutionSegment]
	offset := executionSegment.Len()
	stackSize := uint64(len(stack))
	for idx := uint64(0); idx < stackSize; idx++ {
		if err := executionSegment.Write(offset+uint64(idx), &stack[idx]); err != nil {
			return err
		}
	}

	initialFp := offset + uint64(len(stack)) + cairo1FpOffset
	var err error
	// initialize vm
	runner.vm, err = vm.NewVirtualMachine(vm.Context{
		Pc: *initialPC,
		Ap: initialFp,
		Fp: initialFp,
	}, memory, vm.VirtualMachineConfig{
		ProofMode:    runner.runnerMode == ProofModeCairo0 || runner.runnerMode == ProofModeCairo1,
		CollectTrace: runner.collectTrace,
	})
	return err
}

// run until the program counter equals the `pc` parameter
func (runner *Runner) RunUntilPc(pc *mem.MemoryAddress) error {
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
func (runner *Runner) RunFor(steps uint64) error {
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
func (runner *Runner) EndRun() error {
	for runner.checkUsedCells() != nil {
		pow2Steps := utils.NextPowerOfTwo(runner.vm.Step + 1)
		if err := runner.RunFor(pow2Steps); err != nil {
			return err
		}
	}
	return nil
}

// checkUsedCells returns error if not enough steps were made to allocate required number of cells for builtins
// or there are not enough trace cells to fill the entire range check range
func (runner *Runner) checkUsedCells() error {
	for _, bRunner := range runner.layout.Builtins {
		builtinName := bRunner.Runner.String()
		builtinSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin(builtinName)
		if ok {
			segmentUsedSize := builtinSegment.Len()
			allocatedSize, err := bRunner.Runner.GetAllocatedSize(segmentUsedSize, runner.steps())
			if err != nil {
				return err
			}
			if segmentUsedSize > allocatedSize {
				return fmt.Errorf("builtin %s used size: %d exceeds allocated size: %d ", builtinName, segmentUsedSize, allocatedSize)
			}
		}
	}
	return runner.checkRangeCheckUsage()
}

// Checks if there are not enough trace cells to fill the entire range check range. Each step has assigned a number of range check units. If the number of unused range check units is less than the range of potential values to be checked (defined by rcMin and rcMax), the number of trace cells must be increased, by running additional steps.
func (runner *Runner) checkRangeCheckUsage() error {
	rcMin, rcMax := runner.getPermRangeCheckLimits()
	var rcUnitsUsedByBuiltins uint64
	for _, builtin := range runner.program.Builtins {
		if builtin == builtins.RangeCheckType {
			for _, layoutBuiltin := range runner.layout.Builtins {
				if builtin == layoutBuiltin.Builtin {
					rangeCheckRunner, ok := layoutBuiltin.Runner.(*builtins.RangeCheck)
					if !ok {
						return fmt.Errorf("error type casting to *builtins.RangeCheck")
					}
					rangeCheckSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin(rangeCheckRunner.String())
					if ok {
						rcUnitsUsedByBuiltins += rangeCheckSegment.Len() * rangeCheckRunner.RangeCheckNParts
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
func (runner *Runner) getPermRangeCheckLimits() (uint16, uint16) {
	rcMin, rcMax := runner.vm.RcLimitsMin, runner.vm.RcLimitsMax

	for _, builtin := range runner.program.Builtins {
		if builtin == builtins.RangeCheckType {
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
func (runner *Runner) FinalizeSegments() error {
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

// BuildMemory relocates the memory and returns it
func (runner *Runner) BuildMemory() ([]byte, error) {
	relocatedMemory := runner.vm.RelocateMemory()
	return vm.EncodeMemory(relocatedMemory), nil
}

// BuildTrace relocates the trace and returns it
func (runner *Runner) BuildTrace() ([]byte, error) {
	relocatedTrace := make([]vm.Trace, len(runner.vm.Trace))
	runner.vm.RelocateTrace(&relocatedTrace)
	return vm.EncodeTrace(relocatedTrace), nil
}

func (runner *Runner) pc() mem.MemoryAddress {
	return runner.vm.Context.Pc
}

func (runner *Runner) steps() uint64 {
	return runner.vm.Step
}

// Gives the output of the last run. Panics if there hasn't
// been any runs yet.
func (runner *Runner) Output() []*fp.Element {
	if runner.vm == nil {
		panic("cannot get the output from an uninitialized runner")
	}

	output := []*fp.Element{}
	outputSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin(builtins.OutputName)
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

type InlineCasmContext struct {
	instructions      []*fp.Element
	currentCodeOffset int
}

func (ctx *InlineCasmContext) AddInlineCASM(code string) {
	bytecode, total_size, err := assembler.CasmToBytecode(code)
	if err != nil {
		panic(err)
	}
	ctx.instructions = append(ctx.instructions, bytecode...)
	ctx.currentCodeOffset += int(total_size)
}

func GetEntryCodeInstructions() ([]*fp.Element, error) {
	//TODO: investigate how to implement function param types
	paramTypes := []struct {
		genericTypeId builtins.BuiltinType
		size          int
	}{}
	codeOffset := 0

	ctx := &InlineCasmContext{}

	builtinOffset := map[builtins.BuiltinType]int{
		builtins.PedersenType:     10,
		builtins.RangeCheckType:   9,
		builtins.BitwiseType:      8,
		builtins.ECOPType:         7,
		builtins.PoseidonType:     6,
		builtins.RangeCheck96Type: 5,
		builtins.AddModeType:      4,
		builtins.MulModType:       3,
	}

	emulatedBuiltins := map[builtins.BuiltinType]struct{}{
		1: {},
	}

	apOffset := 0
	paramsSize := 0
	for _, param := range paramTypes {
		ty, size := param.genericTypeId, param.size
		if _, inBuiltin := builtinOffset[ty]; !inBuiltin {
			if _, emulated := emulatedBuiltins[ty]; !emulated && ty != 99 {
				paramsSize += size
			}
		}
	}
	ctx.AddInlineCASM(
		fmt.Sprintf("ap += %d;", paramsSize),
	)
	apOffset += paramsSize

	for _, param := range paramTypes {
		if param.genericTypeId == 99 {
			ctx.AddInlineCASM(
				`%{ memory[ap + 0] = segments.add() %}
				%{ memory[ap + 1] = segments.add() %}
				ap += 2;
				[ap + 0] = 0, ap++;
				[ap - 2] = [[ap - 3]];
				[ap - 1] = [[ap - 3] + 1];
				[ap - 1] = [[ap - 3] + 2];
				apOffset += 3`,
			)
		}
	}

	usedArgs := 0
	for _, param := range paramTypes {
		ty, tySize := param.genericTypeId, param.size
		if offset, isBuiltin := builtinOffset[ty]; isBuiltin {
			ctx.AddInlineCASM(
				fmt.Sprintf("[ap + 0] = [fp - %d], ap++;", offset),
			)
			apOffset += 1
		} else if _, emulated := emulatedBuiltins[ty]; emulated {
			ctx.AddInlineCASM(
				`memory[ap + 0] = segments.add();
				ap += 1;`,
			)
			apOffset += 1
		} else if ty == 99 {
			offset := apOffset - paramsSize
			ctx.AddInlineCASM(
				fmt.Sprintf("[ap + 0] = [ap - %d] + 3, ap++;", offset),
			)
			apOffset += 1
		} else {
			offset := apOffset - usedArgs
			for i := 0; i < tySize; i++ {
				ctx.AddInlineCASM(
					fmt.Sprintf("[ap + 0] = [ap - %d], ap++;", offset),
				)
				apOffset += 1
				usedArgs += 1
			}
		}
	}

	beforeFinalCall := ctx.currentCodeOffset
	finalCallSize := 3
	offset := finalCallSize + codeOffset
	ctx.AddInlineCASM(fmt.Sprintf(`
		call rel %d;
		ret;
	`, offset))
	if beforeFinalCall+finalCallSize != ctx.currentCodeOffset {
		return nil, errors.New("final call offset mismatch")
	}

	return ctx.instructions, nil
}

func GetFooterInstructions() []*fp.Element {
	// Add a `ret` instruction used in libfuncs that retrieve the current value of the `fp`
	// and `pc` registers.
	return []*fp.Element{new(fp.Element).SetUint64(2345108766317314046)}
}
