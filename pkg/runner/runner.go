package runner

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/NethermindEth/cairo-vm-go/pkg/assembler"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type RunnerMode uint8

const (
	ExecutionModeZero RunnerMode = iota + 1
	ProofModeZero
	ExecutionModeCairo
	ProofModeCairo
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
func NewRunner(program *Program, hints map[uint64][]hinter.Hinter, runnerMode RunnerMode, collectTrace bool, maxsteps uint64, layoutName string, userArgs []starknet.CairoFuncArgs, availableGas uint64) (Runner, error) {
	layout, err := builtins.GetLayout(layoutName)
	if err != nil {
		return Runner{}, err
	}
	newHintRunnerContext := getNewHintRunnerContext(program, userArgs, availableGas, runnerMode == ProofModeCairo || runnerMode == ProofModeZero)
	hintrunner := hintrunner.NewHintRunner(hints, &newHintRunnerContext)
	return Runner{
		program:      program,
		runnerMode:   runnerMode,
		hintrunner:   hintrunner,
		collectTrace: collectTrace,
		maxsteps:     maxsteps,
		layout:       layout,
	}, nil
}

func getNewHintRunnerContext(program *Program, userArgs []starknet.CairoFuncArgs, availableGas uint64, proofmode bool) hinter.HintRunnerContext {
	// The writeApOffset is the offset where the user arguments will be written. It is added to the current AP in the ExternalWriteArgsToMemory hint.
	// The writeApOffset is significant for Cairo programs, because of the prepended Entry Code instructions.
	// In the entry code instructions the builtins bases (excluding gas, segment arena and output) are written to the memory,
	// thus the writeApOffset should be increased by the number of builtins.
	// In the entry code the instructions for programs utilizing the SegmentArena are also prepended. The SegmentArena is a builtin that requires 4 cells:
	//  * segment_arena_ptr
	//  * info_segment_ptr
	//  * 0
	//  * segment_arena_ptr + 3
	// But the builtin itself shouldn't be included in len(builtins), therefore the writeApOffset should be increased by 3.
	writeApOffset := uint64(len(program.Builtins))
	for _, builtin := range program.Builtins {
		if builtin == builtins.SegmentArenaType {
			writeApOffset += 3
		}
	}
	if proofmode {
		writeApOffset += uint64(len(program.Builtins)) - 1
	}

	newHintrunnerContext := *hinter.InitializeDefaultContext()
	err := newHintrunnerContext.ScopeManager.AssignVariables(map[string]any{
		"userArgs":             userArgs,
		"apOffset":             writeApOffset,
		"gas":                  availableGas,
		"useTemporarySegments": proofmode,
	})
	// Error handling: this condition should never be true, since the context was initialized above
	if err != nil {
		panic(fmt.Sprintf("assign variables: %v", err))
	}
	return newHintrunnerContext
}

func AssembleProgram(cairoProgram *starknet.StarknetProgram, userArgs []starknet.CairoFuncArgs, availableGas uint64, proofmode bool) (Program, map[uint64][]hinter.Hinter, []starknet.CairoFuncArgs, error) {
	mainFunc, ok := cairoProgram.EntryPointsByFunction["main"]
	if !ok {
		return Program{}, nil, nil, fmt.Errorf("cannot find main function")
	}

	if proofmode {
		err := CheckOnlyArrayFeltInputAndReturntValue(mainFunc)
		if err != nil {
			return Program{}, nil, nil, err
		}
	}

	expectedArgsSize, actualArgsSize := 0, 0
	for _, arg := range mainFunc.InputArgs {
		expectedArgsSize += arg.Size
	}
	for _, arg := range userArgs {
		if arg.Single != nil {
			actualArgsSize += 1
		} else {
			actualArgsSize += 2
		}
	}
	if expectedArgsSize != actualArgsSize {
		return Program{}, nil, nil, fmt.Errorf("missing arguments for main function")
	}
	program, err := LoadCairoProgram(cairoProgram)
	if err != nil {
		return Program{}, nil, nil, fmt.Errorf("cannot load program: %w", err)
	}
	hints, err := core.GetCairoHints(cairoProgram)
	if err != nil {
		return Program{}, nil, nil, fmt.Errorf("cannot get hints: %w", err)
	}

	entryCodeInstructions, entryCodeHints, err := GetEntryCodeInstructions(mainFunc, proofmode)

	if err != nil {
		return Program{}, nil, nil, fmt.Errorf("cannot load entry code instructions: %w", err)
	}
	program.Bytecode = append(entryCodeInstructions, program.Bytecode...)
	program.Bytecode = append(program.Bytecode, GetFooterInstructions()...)

	offset := uint64(len(entryCodeInstructions))
	shiftedHintsMap := make(map[uint64][]hinter.Hinter)
	for key, value := range hints {
		shiftedHintsMap[key+offset] = value
	}
	for key, hint := range entryCodeHints {
		shiftedHintsMap[key] = hint
	}
	return *program, shiftedHintsMap, userArgs, nil
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
	if runner.runnerMode == ProofModeCairo {
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

	if runner.isProofMode() {
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
	case ExecutionModeZero, ExecutionModeCairo, ProofModeCairo:
		returnFp := memory.AllocateEmptySegment()
		mvReturnFp := mem.MemoryValueFromMemoryAddress(&returnFp)
		if runner.runnerMode == ProofModeCairo {
			// In Cairo mainPCOffset is equal to the offset of program segment base
			return runner.initializeEntrypoint(uint64(0), nil, &mvReturnFp, memory, stack, 2)
		} else if runner.runnerMode == ExecutionModeCairo {
			return runner.initializeEntrypoint(uint64(0), nil, &mvReturnFp, memory, stack, 0)
		} else {
			mainPCOffset, ok := runner.program.Entrypoints["main"]
			if !ok {
				return mem.UnknownAddress, errors.New("can't find an entrypoint for main")
			}
			return runner.initializeEntrypoint(mainPCOffset, nil, &mvReturnFp, memory, stack, 0)
		}
	case ProofModeZero:
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
		if programBuiltin == builtins.GasBuiltinType || programBuiltin == builtins.SegmentArenaType {
			continue
		}
		if _, found := builtinsSet[programBuiltin]; !found {
			builtinName, err := programBuiltin.MarshalJSON()
			if err != nil {
				return []mem.MemoryValue{}, err
			}
			return []mem.MemoryValue{}, fmt.Errorf("builtin %s not found in the layout: %s", builtinName, runner.layout.Name)
		}
	}
	stack := []mem.MemoryValue{}

	for _, bRunner := range runner.layout.Builtins {
		if runner.isCairoMode() {
			if utils.Contains(runner.program.Builtins, bRunner.Builtin) {
				builtinSegment := memory.AllocateBuiltinSegment(bRunner.Runner)
				stack = append(stack, mem.MemoryValueFromMemoryAddress(&builtinSegment))
			}
		} else {
			builtinSegment := memory.AllocateBuiltinSegment(bRunner.Runner)
			if utils.Contains(runner.program.Builtins, bRunner.Builtin) {
				stack = append(stack, mem.MemoryValueFromMemoryAddress(&builtinSegment))
			}
		}
	}
	// Write builtins costs segment address to the end of the program segment
	if runner.isCairoMode() {
		err := gasInitialization(memory)
		if err != nil {
			return nil, err
		}
	}
	return stack, nil
}

func (runner *Runner) isProofMode() bool {
	return runner.runnerMode == ProofModeCairo || runner.runnerMode == ProofModeZero
}

func (runner *Runner) isCairoMode() bool {
	return runner.runnerMode == ExecutionModeCairo || runner.runnerMode == ProofModeCairo
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
		ProofMode:    runner.isProofMode(),
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

// Function derived from the cairo-lang-runner crate.
// https://github.com/starkware-libs/cairo/blob/40a7b60687682238f7f71ef7c59c986cc5733915/crates/cairo-lang-runner/src/lib.rs#L703
// / Returns the instructions to add to the beginning of the code to successfully call the main
// / function, as well as the builtins required to execute the program.
func GetEntryCodeInstructions(function starknet.EntryPointByFunction, proofmode bool) ([]*fp.Element, map[uint64][]hinter.Hinter, error) {
	paramTypes := function.InputArgs
	apOffset := 0
	builtinOffset := 3
	codeOffset := uint64(function.Offset)
	builtinsOffsetsMap := map[builtins.BuiltinType]int{}
	ctx := &InlineCasmContext{}

	for _, builtin := range []builtins.BuiltinType{
		builtins.MulModType,
		builtins.AddModeType,
		builtins.RangeCheck96Type,
		builtins.PoseidonType,
		builtins.ECOPType,
		builtins.BitwiseType,
		builtins.RangeCheckType,
		builtins.PedersenType,
	} {
		if slices.Contains(function.Builtins, builtin) {
			builtinsOffsetsMap[builtins.BuiltinType(builtin)] = builtinOffset
			builtinOffset += 1
		}
	}

	if proofmode {
		builtinsOffsetsMap[builtins.OutputType] = 0
		// Increment the length of the builtins by 1 to account for the output builtin
		ctx.AddInlineCASM(fmt.Sprintf("ap += %d;", len(function.Builtins)+1))
	}

	gotSegmentArena := false
	for _, builtin := range function.Builtins {
		if builtin == builtins.SegmentArenaType {
			gotSegmentArena = true
		}
	}

	hints := make(map[uint64][]hinter.Hinter)

	if gotSegmentArena {
		hints[uint64(ctx.currentCodeOffset)] = []hinter.Hinter{
			&core.AllocSegment{
				Dst: hinter.ApCellRef(0),
			},
			&core.AllocSegment{
				Dst: hinter.ApCellRef(1),
			},
		}
		ctx.AddInlineCASM(
			"[ap+2] = 0, ap++;",
		)
		ctx.AddInlineCASM(
			"[ap] = [[ap-1]], ap++;",
		)
		ctx.AddInlineCASM(
			`
			[ap] = [[ap-2]+1], ap++;
			[ap-1] = [[ap-3]+2];
			`,
		)
		apOffset += 3
	}

	paramsSize := 0
	for _, param := range paramTypes {
		paramsSize += param.Size
	}
	apOffset += paramsSize
	gotGasBuiltin := false

	for _, builtin := range function.Builtins {
		if offset, isBuiltin := builtinsOffsetsMap[builtin]; isBuiltin {
			ctx.AddInlineCASM(
				fmt.Sprintf("[ap + 0] = [fp - %d], ap++;", offset),
			)
			apOffset += 1
		} else if builtin == builtins.SegmentArenaType {
			offset := apOffset - paramsSize
			ctx.AddInlineCASM(
				fmt.Sprintf("[ap + 0] = [ap - %d] + 3, ap++;", offset),
			)
			apOffset += 1
		} else if builtin == builtins.GasBuiltinType {
			hints[uint64(ctx.currentCodeOffset)] = append(hints[uint64(ctx.currentCodeOffset)], &core.ExternalWriteGasToMemory{})
			ctx.AddInlineCASM("ap += 1;")
			apOffset += 1
			gotGasBuiltin = true
		}
	}

	// Incrementing the AP for the input args, because their values are written to memory by the VM in the ExternalWriteArgsToMemory hint.
	for _, param := range paramTypes {
		ctx.AddInlineCASM(
			fmt.Sprintf("ap+=%d;", param.Size),
		)
	}

	// The hint can be executed before the first instruction, because the AP correction was calculated based on the input arguments.
	if paramsSize > 0 {
		hints[uint64(0)] = append(hints[uint64(0)], []hinter.Hinter{
			&core.ExternalWriteArgsToMemory{},
		}...)
	}

	_, endInstructionsSize, err := assembler.CasmToBytecode("call rel 0; ret;")
	if err != nil {
		return nil, nil, err
	}
	totalSize := uint64(endInstructionsSize) + uint64(codeOffset)
	ctx.AddInlineCASM(fmt.Sprintf("call rel %d; ret;", int(totalSize)))
	outputPtr := fmt.Sprintf("[fp-%d]", len(function.Builtins)+2)
	if proofmode {
		for i, b := range function.Builtins {
			// assert [fp + i] == [fp - builtin_offset]
			offset := builtinsOffsetsMap[b]
			// increment the offset by 1 to account for the skipped output builtin
			ctx.AddInlineCASM(fmt.Sprintf("[fp+%d] = [ap-%d];", i+1, offset))
		}
		outputs := []string{}

		lastReturnArg := function.ReturnArgs[len(function.ReturnArgs)-1]
		for i := function.ReturnArgs[len(function.ReturnArgs)-1].Size + 1; i >= 1; i-- {
			outputs = append(outputs, fmt.Sprintf("[ap-%d]", i))
		}

		arrayStartPtr, arrayEndPtr := outputs[0], outputs[1]
		if strings.HasPrefix(lastReturnArg.DebugName, "core::panics::PanicResult") {
			// assert panic_flag = *(output_ptr++);
			panicFlag := outputs[0]
			ctx.AddInlineCASM(fmt.Sprintf("%s = [%s];", panicFlag, outputPtr))
			arrayStartPtr, arrayEndPtr = outputs[1], outputs[2]
		}

		ctx.AddInlineCASM(
			fmt.Sprintf(`
				%s = [ap] + %s, ap++;
				[ap-1] = [%s+1];
				[ap] = [ap-1], ap++;
				[ap] = %s, ap++;
				[ap] = %s + 1, ap++;
				jmp rel 4 if [ap-3] != 0;
				jmp rel 12;

				[ap] = [[ap-2]], ap++;
				[ap-1] = [[ap-2]];
				[ap-4] = [ap] + 1, ap++;
				[ap] = [ap-4] + 1, ap++;
				[ap] = [ap-4] + 1, ap++;
				jmp rel -8 if [ap-3] != 0;
			`, arrayEndPtr, arrayStartPtr, outputPtr, arrayStartPtr, outputPtr),
		)
		if paramsSize != 0 {
			offset := 2 * len(function.Builtins)
			if gotSegmentArena {
				offset += 4
			}
			if gotGasBuiltin {
				offset += 1
			}
			arrayStartPtr := fmt.Sprintf("[fp+%d]", offset)
			arrayEndPtr := fmt.Sprintf("[fp+%d]", offset+1)

			ctx.AddInlineCASM(
				fmt.Sprintf(`
					%s = [ap] + %s, ap++;
					[ap-1] = [[ap-2]];
					[ap] = [ap-1], ap++;
					[ap] = %s, ap++;
					[ap] = [ap-4]+1, ap++;
					jmp rel 4 if [ap-3] != 0;
					jmp rel 12;

					[ap] = [[ap-2]], ap++;
					[ap-1] = [[ap-2]];
					[ap-4] = [ap]+1, ap++;
					[ap] = [ap-4]+1, ap++;
					[ap] = [ap-4]+1;
					jmp rel -8 if [ap-3] != 0;
				`, arrayEndPtr, arrayStartPtr, arrayStartPtr),
			)
		}
		// After we are done writing into the output segment, we can write the final output_ptr into locals:
		// The last instruction will write the final output ptr so we can find it in [ap - 1]
		ctx.AddInlineCASM("[fp] = [[ap - 1]];")

		if gotSegmentArena {
			offset := 2 + len(function.Builtins)
			segmentArenaPtr := fmt.Sprintf("[fp + %d]", offset)
			hints[uint64(ctx.currentCodeOffset)] = append(hints[uint64(ctx.currentCodeOffset)], &core.RelocateAllDictionaries{})
			ctx.AddInlineCASM(fmt.Sprintf(`
				[ap]=[%s-2], ap++;
				[ap]=[%s-1], ap++;
				[ap-1]=[ap-2];
				jmp rel 4 if [ap-2] != 0;
				jpm rel 19;
				[ap]=[[%s-3], ap++;
				[ap-3] = [ap]+1;
				jmp rel 4 if [ap-1] != 0;
				jmp rel 12;
				[ap]=[[ap-2]+1], ap++;
				[ap] = [[ap-3]+3], ap++;
        		[ap-1] = [ap-2] + 1;
        		[ap] = [ap-4] + 3, ap++;
        		[ap-4] = [ap] + 1, ap++;
        		jmp rel -12;
				`, segmentArenaPtr, segmentArenaPtr, segmentArenaPtr,
			))
		}

		// Copying the final builtins from locals into the top of the stack.
		for i := range function.Builtins {
			ctx.AddInlineCASM(fmt.Sprintf("[ap] = [fp + %d], ap++;", i))
		}
	} else {
		// Writing the final builtins into the top of the stack.
		for _, b := range function.Builtins {
			offset := builtinsOffsetsMap[b]
			ctx.AddInlineCASM(fmt.Sprintf("[ap] = [fp - %d], ap++;", offset))
		}

	}
	if proofmode {
		ctx.AddInlineCASM("jmp rel 0;")
	} else {
		ctx.AddInlineCASM("ret;")
	}
	return ctx.instructions, hints, nil
}

func GetFooterInstructions() []*fp.Element {
	// Add a `ret` instruction used in libfuncs that retrieve the current value of the `fp`
	// and `pc` registers.
	return []*fp.Element{new(fp.Element).SetUint64(2345108766317314046)}
}

func (runner *Runner) FinalizeBuiltins() error {
	// Finalization of builtins is done only in proofmode with air public input
	// It could also be implemented in execution mode, if cairo pie output was
	// implemented.
	if runner.runnerMode == ProofModeCairo {
		builtinNameToStackPointer := map[builtins.BuiltinType]uint64{}
		for i, builtin := range runner.program.Builtins {
			builtinNameToStackPointer[builtin] = runner.vm.Context.Ap - uint64(len(runner.program.Builtins)-i-1)
		}
		err := runner.vm.BuiltinsFinalStackFromStackPointerDict(builtinNameToStackPointer)
		if err != nil {
			return err
		}
	}
	return nil
}

func CheckOnlyArrayFeltInputAndReturntValue(mainFunc starknet.EntryPointByFunction) error {
	if len(mainFunc.InputArgs) != 1 {
		return fmt.Errorf("main function in proofmode should have felt252 array as input argument")
	}
	if len(mainFunc.ReturnArgs) != 1 {
		return fmt.Errorf("main function in proofmode should have an felt252 array as return argument")
	}
	if mainFunc.InputArgs[0].Size != 2 || mainFunc.InputArgs[0].DebugName != "Array<felt252>" {
		return fmt.Errorf("main function input argument should be Felt Array")
	}
	if mainFunc.ReturnArgs[0].Size != 3 || mainFunc.ReturnArgs[0].DebugName != "core::panics::PanicResult::<(core::array::Array::<core::felt252>)>" {
		return fmt.Errorf("main function return argument should be Felt Array")
	}
	return nil
}

func (runner *Runner) GetAirMemorySegmentsAddresses() (map[string]AirMemorySegmentEntry, error) {
	segmentsOffsets, _ := runner.vm.Memory.RelocationOffsets()
	memorySegmentsAddresses := make(map[string]AirMemorySegmentEntry)
	for segmentIndex, segment := range runner.vm.Memory.Segments {
		if segment.BuiltinRunner == nil {
			continue
		}
		if segmentIndex >= len(segmentsOffsets) {
			return nil, fmt.Errorf("segment index %d not found in segments offsets", segmentIndex)
		}
		bRunner := segment.BuiltinRunner
		stopPtr := bRunner.GetStopPointer()
		baseOffset := segmentsOffsets[segmentIndex]
		memorySegmentsAddresses[bRunner.String()] = AirMemorySegmentEntry{BeginAddr: baseOffset, StopPtr: baseOffset + stopPtr}
	}
	return memorySegmentsAddresses, nil
}
