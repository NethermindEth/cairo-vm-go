package runner

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
)

func (runner *ZeroRunner) GetAirPublicInput() (AirPublicInput, error) {
	rcMin, rcMax := runner.getPermRangeCheckLimits()

	// TODO: refactor to reuse earlier computed relocated trace
	relocatedTrace := runner.vm.RelocateTrace()
	firstTrace := relocatedTrace[0]
	lastTrace := relocatedTrace[len(relocatedTrace)-1]
	memorySegments := make(map[string]AirMemorySegmentEntry)
	// TODO: you need to calculate this for each builtin
	memorySegments["program"] = AirMemorySegmentEntry{BeginAddr: firstTrace.Pc, StopPtr: lastTrace.Pc}
	memorySegments["execution"] = AirMemorySegmentEntry{BeginAddr: firstTrace.Ap, StopPtr: lastTrace.Ap}

	return AirPublicInput{
		Layout:        runner.layout.Name,
		RcMin:         rcMin,
		RcMax:         rcMax,
		NSteps:        len(runner.vm.Trace),
		DynamicParams: nil,
		// TODO: yet to be implemented fully
		MemorySegments: memorySegments,
		// TODO: yet to be implemented
		PublicMemory: make([]AirPublicMemoryEntry, 0),
	}, nil
}

type AirPublicInput struct {
	Layout         string                           `json:"layout"`
	RcMin          uint16                           `json:"rc_min"`
	RcMax          uint16                           `json:"rc_max"`
	NSteps         int                              `json:"n_steps"`
	DynamicParams  interface{}                      `json:"dynamic_params"`
	MemorySegments map[string]AirMemorySegmentEntry `json:"memory_segments"`
	PublicMemory   []AirPublicMemoryEntry           `json:"public_memory"`
}

type AirMemorySegmentEntry struct {
	BeginAddr uint64 `json:"begin_addr"`
	StopPtr   uint64 `json:"stop_ptr"`
}

type AirPublicMemoryEntry struct {
	Address uint16 `json:"address"`
	Value   string `json:"value"`
	Page    uint16 `json:"page"`
}

func (runner *ZeroRunner) GetAirPrivateInput(tracePath, memoryPath string) (AirPrivateInput, error) {
	airPrivateInput := AirPrivateInput{
		TracePath:  tracePath,
		MemoryPath: memoryPath,
	}

	for _, bRunner := range runner.layout.Builtins {
		builtinName := bRunner.Runner.String()
		builtinSegment, ok := runner.vm.Memory.FindSegmentWithBuiltin(builtinName)
		if ok {
			// some checks might be missing here
			switch builtinName {
			case builtins.RangeCheckName:
				{
					airPrivateInput.RangeCheck = bRunner.Runner.(*builtins.RangeCheck).GetAirPrivateInput(builtinSegment)
				}
			case builtins.BitwiseName:
				{
					airPrivateInput.Bitwise = bRunner.Runner.(*builtins.Bitwise).GetAirPrivateInput(builtinSegment)
				}
			case builtins.PoseidonName:
				{
					airPrivateInput.Poseidon = bRunner.Runner.(*builtins.Poseidon).GetAirPrivateInput(builtinSegment)
				}
			case builtins.PedersenName:
				{
					airPrivateInput.Pedersen = bRunner.Runner.(*builtins.Pedersen).GetAirPrivateInput(builtinSegment)
				}
			case builtins.EcOpName:
				{
					airPrivateInput.EcOp = bRunner.Runner.(*builtins.EcOp).GetAirPrivateInput(builtinSegment)
				}
			case builtins.KeccakName:
				{
					airPrivateInput.Keccak = bRunner.Runner.(*builtins.Keccak).GetAirPrivateInput(builtinSegment)
				}
			case builtins.ECDSAName:
				{
					ecdsaAirPrivateInput, err := bRunner.Runner.(*builtins.ECDSA).GetAirPrivateInput(builtinSegment)
					if err != nil {
						return AirPrivateInput{}, err
					}
					airPrivateInput.Ecdsa = ecdsaAirPrivateInput
				}
			}
		}
	}

	return airPrivateInput, nil
}

type AirPrivateInput struct {
	TracePath  string                                 `json:"trace_path"`
	MemoryPath string                                 `json:"memory_path"`
	Pedersen   []builtins.AirPrivateBuiltinPedersen   `json:"pedersen"`
	RangeCheck []builtins.AirPrivateBuiltinRangeCheck `json:"range_check"`
	Ecdsa      []builtins.AirPrivateBuiltinECDSA      `json:"ecdsa"`
	Bitwise    []builtins.AirPrivateBuiltinBitwise    `json:"bitwise"`
	EcOp       []builtins.AirPrivateBuiltinEcOp       `json:"ec_op"`
	Keccak     []builtins.AirPrivateBuiltinKeccak     `json:"keccak"`
	Poseidon   []builtins.AirPrivateBuiltinPoseidon   `json:"poseidon"`
}
