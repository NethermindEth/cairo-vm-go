package zero

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner"
	"github.com/NethermindEth/cairo-vm-go/pkg/safemath"
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
func NewRunner(program *Program, proofmode bool, maxsteps uint64) (*ZeroRunner, error) {
	// todo(rodro): given the program get the appropiate hints
	hintrunner := hintrunner.NewHintRunner(make(map[uint64]hintrunner.Hinter))

	return &ZeroRunner{
		program:    program,
		hintrunner: hintrunner,
		proofmode:  proofmode,
		maxsteps:   maxsteps,
	}, nil
}

// todo(rodro): should we add support for running any function?
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
		pow2Steps := safemath.NextPowerOfTwo(runner.vm.Step + 1)
		if err := runner.RunFor(pow2Steps); err != nil {
			return err
		}
	}
	return nil
}

func (runner *ZeroRunner) InitializeMainEntrypoint() (mem.MemoryAddress, error) {
	memory := mem.InitializeEmptyMemory()
	_, err := memory.AllocateSegment(runner.program.Bytecode) // ProgramSegment
	if err != nil {
		return mem.UnknownAddress, err
	}

	memory.AllocateEmptySegment() // ExecutionSegment
	if runner.proofmode {
		initialPCOffset, ok := runner.program.Labels["__start__"]
		if !ok {
			return mem.UnknownAddress, errors.New("start label not found. Try compiling with `--proof_mode`")
		}
		endPcOffset, ok := runner.program.Labels["__end__"]
		if !ok {
			return mem.UnknownAddress, errors.New("end label not found. Try compiling with `--proof_mode`")
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

	returnFp := mem.MemoryValueFromSegmentAndOffset(
		memory.AllocateEmptySegment(),
		0,
	)
	return runner.InitializeEntrypoint("main", nil, &returnFp, memory)
}

func (runner *ZeroRunner) InitializeEntrypoint(
	funcName string, arguments []*f.Element, returnFp *mem.MemoryValue, memory *mem.Memory,
) (mem.MemoryAddress, error) {
	initialPCOffset, ok := runner.program.Entrypoints[funcName]
	if !ok {
		return mem.UnknownAddress, fmt.Errorf("unknown entrypoint: %s", funcName)
	}

	stack := runner.initializeBuiltins(memory)
	for i := range arguments {
		stack = append(stack, mem.MemoryValueFromFieldElement(arguments[i]))
	}
	end := mem.MemoryAddress{
		SegmentIndex: uint64(memory.AllocateEmptySegment()),
		Offset:       0,
	}

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
		stack = append(stack, mem.MemoryValueFromSegmentAndOffset(builtinSegment, 0))
	}
	return stack
}

func (runner *ZeroRunner) initializeVm(initialPC *mem.MemoryAddress, stack []mem.MemoryValue, memory *mem.Memory) error {
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
		if err := runner.vm.RunStep(runner.hintrunner); err != nil {
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
		if err := runner.vm.RunStep(runner.hintrunner); err != nil {
			return fmt.Errorf(
				"pc %s step %d: %w",
				runner.pc().String(),
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

	return EncodeTrace(relocatedTrace), EncodeMemory(runner.vm.RelocateMemory()), nil
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
	for _, segment := range runner.vm.Memory.Segments {
		if segment.BuiltinRunner.String() == "output" {
			for offset := uint64(0); offset < segment.Len(); offset++ {
				value := segment.Peek(offset)
				// todo(rodro): check if output can only contains field elements
				valueFelt, _ := value.FieldElement()
				output = append(output, valueFelt)
			}
			break
		}
	}
	return output
}

const ctxSize = 3 * 8

func EncodeTrace(trace []vm.Trace) []byte {
	content := make([]byte, 0, len(trace)*ctxSize)
	for i := range trace {
		content = binary.LittleEndian.AppendUint64(content, trace[i].Ap)
		content = binary.LittleEndian.AppendUint64(content, trace[i].Fp)
		content = binary.LittleEndian.AppendUint64(content, trace[i].Pc)
	}
	return content
}

func DecodeTrace(content []byte) []vm.Trace {
	trace := make([]vm.Trace, 0, len(content)/ctxSize)
	for i := 0; i < len(content); i += ctxSize {
		trace = append(
			trace,
			vm.Trace{
				Ap: binary.LittleEndian.Uint64(content[i : i+8]),
				Fp: binary.LittleEndian.Uint64(content[i+8 : i+16]),
				Pc: binary.LittleEndian.Uint64(content[i+16 : i+24]),
			},
		)
	}
	return trace
}

const addrSize = 8
const feltSize = 32

// Encody the relocated memory in the (address, value) form
// in a consecutive way
func EncodeMemory(memory []*f.Element) []byte {
	// Check non nil elements for optimal array size
	nonNilElms := 0
	for i := range memory {
		if memory[i] != nil {
			nonNilElms++
		}
	}
	content := make([]byte, nonNilElms*(addrSize+feltSize))

	count := 0
	for i := range memory {
		if memory[i] == nil {
			continue
		}
		// set the right content index
		j := count * (addrSize + feltSize)
		// store the address
		binary.LittleEndian.PutUint64(content[j:j+addrSize], uint64(i))
		// store the field element
		f.LittleEndian.PutElement(
			(*[32]byte)(content[j+addrSize:j+addrSize+feltSize]),
			*memory[i],
		)

		// increase the number of elements stored
		count++
	}
	return content
}

func DecodeMemory(content []byte) []*f.Element {
	// calculate the max memory index
	lastContentInd := len(content) - (addrSize + feltSize)
	lasMemIndex := binary.LittleEndian.Uint64(content[lastContentInd : lastContentInd+addrSize])

	// create the memory array with the same length as the max memory index
	memory := make([]*f.Element, lasMemIndex+1)

	// decode the encontent and store it in memory
	for i := 0; i < len(content); i += addrSize + feltSize {
		memIndex := binary.LittleEndian.Uint64(content[i : i+addrSize])
		felt, err := f.LittleEndian.Element((*[32]byte)(content[i+addrSize : i+addrSize+feltSize]))
		if err != nil {
			panic(err)
		}
		memory[memIndex] = &felt
	}
	return memory
}
