package zero

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Program struct {
	// the bytecode in string format
	Bytecode []*f.Element
	// given a string it returns the pc for that function call
	Entrypoints map[string]uint64
	// it stores the start and end label pcs
	Labels map[string]uint64
}

func LoadCairoZeroProgram(content []byte) (*Program, error) {
	cairoZeroJson, err := zero.ZeroProgramFromJSON(content)
	if err != nil {
		return nil, err
	}

	// bytecode
	bytecode := make([]*f.Element, len(cairoZeroJson.Data))
	for i := range cairoZeroJson.Data {
		felt, err := new(f.Element).SetString(cairoZeroJson.Data[i])
		if err != nil {
			return nil, fmt.Errorf(
				"cannot read bytecode %s at position %d: %w",
				cairoZeroJson.Data[i], i, err,
			)
		}
		bytecode[i] = felt
	}

	entrypoints, err := extractEntrypoints(cairoZeroJson)
	if err != nil {
		return nil, err
	}

	labels, err := extractLabels(cairoZeroJson)
	if err != nil {
		return nil, err
	}

	return &Program{
		Bytecode:    bytecode,
		Entrypoints: entrypoints,
		Labels:      labels,
	}, nil
}

func extractEntrypoints(json *zero.ZeroProgram) (map[string]uint64, error) {
	result := make(map[string]uint64)
	err := scanIdentifiers(
		json,
		func(key string, typex string, value map[string]any) error {
			if typex == "function" {
				pc, ok := value["pc"].(float64)
				if !ok {
					return fmt.Errorf("%s: unknown entrypoint pc", key)
				}
				name := key[len(json.MainScope)+1:]
				result[name] = uint64(pc)
			}
			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("extracting entrypoints: %w", err)
	}
	return result, nil
}

func extractLabels(json *zero.ZeroProgram) (map[string]uint64, error) {
	labels := make(map[string]uint64, 2)
	err := scanIdentifiers(
		json,
		func(key string, typex string, value map[string]any) error {
			if typex == "label" {
				pc, ok := value["pc"].(float64)
				if !ok {
					return fmt.Errorf("%s: unknown entrypoint pc", key)
				}
				name := key[len(json.MainScope)+1:]
				labels[name] = uint64(pc)
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("extracting labels: %w", err)
	}

	return labels, nil
}

func scanIdentifiers(
	json *zero.ZeroProgram,
	f func(key string, typex string, value map[string]any) error,
) error {
	for key, value := range json.Identifiers {
		properties := value.(map[string]any)

		typex, ok := properties["type"].(string)
		if !ok {
			return errors.New("unnespecified identifier type")
		}
		if err := f(key, typex, properties); err != nil {
			return err
		}
	}
	return nil
}

type ZeroRunner struct {
	// core components
	program    *Program
	vm         *VM.VirtualMachine
	hintrunner hintrunner.HintRunner
	// config
	proofmode bool
	// auxiliar
	runFinished bool
}

// Creates a new Runner of a Cairo Zero program
func NewRunner(program *Program, proofmode bool) (*ZeroRunner, error) {

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

// todo(rodro): should we add support for running any function?
func (runner *ZeroRunner) Run() error {
	if runner.runFinished {
		return errors.New("cannot re-run using the same runner")
	}

	end, err := runner.InitializeMainEntrypoint()
	if err != nil {
		return fmt.Errorf("initializing main entry point: %w", err)
	}

	err = runner.RunUntilPc(end)
	if err != nil {
		return fmt.Errorf("step %d, pc %d:\n%w", runner.vm.Step, runner.vm.Context.Pc, err)
	}
	return nil
}

func (runner *ZeroRunner) InitializeMainEntrypoint() (uint64, error) {
	if runner.proofmode {
		startPc, ok := runner.program.Labels["__start__"]
		if !ok {
			return 0, errors.New("start label not found. Try compiling with `--proof_mode`")
		}
		endPc, ok := runner.program.Labels["__end__"]
		if !ok {
			return 0, errors.New("end label not found. Try compiling with `--proof_mode`")
		}

		offset := runner.segments()[VM.ExecutionSegment].Len()

		// set dummy fp value
		err := runner.memory().Write(
			VM.ExecutionSegment,
			offset,
			memory.MemoryValueFromSegmentAndOffset(VM.ExecutionSegment, offset+2),
		)
		if err != nil {
			return 0, err
		}
		// set dummy pc value
		err = runner.memory().Write(VM.ExecutionSegment, offset+1, memory.MemoryValueFromUint[uint64](0))
		if err != nil {
			return 0, err
		}

		runner.vm.Context.Pc = startPc
		runner.vm.Context.Ap = offset + 2
		runner.vm.Context.Fp = runner.vm.Context.Ap
		return endPc, nil
	}

	returnFp := memory.MemoryValueFromSegmentAndOffset(
		runner.memory().AllocateEmptySegment(),
		0,
	)
	return runner.InitializeEntrypoint("main", nil, returnFp)
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
	err = runner.memory().Write(VM.ExecutionSegment, offset+1, memory.MemoryValueFromUint(end))
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

// todo(rodro):
// 3. Prepare the runner to run in proof mode
// 5. Use fibonacci test
// 6. Debug accordingly
// 7. Compare trace with py-vm and the rust-vm trace
func (runner *ZeroRunner) BuildProof() ([]byte, []byte, error) {
	relocatedTrace, relocatedMem, err := runner.vm.Proof()
	if err != nil {
		return nil, nil, err
	}

	trace := EncodeTrace(relocatedTrace)
	memory := EncodeMemory(relocatedMem)

	return trace, memory, nil
}

func (runner *ZeroRunner) memory() *memory.Memory {
	return runner.vm.MemoryManager.Memory
}

func (runner *ZeroRunner) segments() []*memory.Segment {
	return runner.vm.MemoryManager.Memory.Segments
}

const ctxSize = 3 * 8

func EncodeTrace(trace []vm.Context) []byte {
	content := make([]byte, 0, len(trace)*ctxSize)
	for i := range trace {
		content = binary.LittleEndian.AppendUint64(content, trace[i].Ap)
		content = binary.LittleEndian.AppendUint64(content, trace[i].Fp)
		content = binary.LittleEndian.AppendUint64(content, trace[i].Pc)
	}
	return content
}

func DecodeTrace(content []byte) []vm.Context {
	trace := make([]vm.Context, 0, len(content)/ctxSize)
	for i := 0; i < len(content); i += ctxSize {
		trace = append(
			trace,
			VM.Context{
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
