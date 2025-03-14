package runner

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func (runner *Runner) GetAirPublicInput(relocatedMemory []*fp.Element, publicMemoryAddresses []vm.PublicMemoryAddress) (AirPublicInput, error) {
	rcMin, rcMax := runner.getPermRangeCheckLimits()

	// TODO: refactor to reuse earlier computed relocated trace
	relocatedTrace := make([]vm.Trace, len(runner.vm.Trace))
	runner.vm.RelocateTrace(&relocatedTrace)
	firstTrace := relocatedTrace[0]
	lastTrace := relocatedTrace[len(relocatedTrace)-1]
	memorySegments := make(map[string]AirMemorySegmentEntry)
	memorySegments["program"] = AirMemorySegmentEntry{BeginAddr: firstTrace.Pc, StopPtr: lastTrace.Pc}
	memorySegments["execution"] = AirMemorySegmentEntry{BeginAddr: firstTrace.Ap, StopPtr: lastTrace.Ap}
	memorySegmentsAddresses, err := runner.GetAirMemorySegmentsAddresses()
	if err != nil {
		return AirPublicInput{}, err
	}
	for name, segment := range memorySegmentsAddresses {
		memorySegments[name] = segment
	}
	publicMemory := make([]AirPublicMemoryEntry, len(publicMemoryAddresses))

	for i, address := range publicMemoryAddresses {
		publicMemory[i] = AirPublicMemoryEntry{
			Address: address.Address,
			Page:    address.Page,
			Value:   "0x" + relocatedMemory[address.Address].Text(16),
		}
	}

	return AirPublicInput{
		Layout:         runner.layout.Name,
		RcMin:          rcMin,
		RcMax:          rcMax,
		NSteps:         len(runner.vm.Trace),
		DynamicParams:  nil,
		MemorySegments: memorySegments,
		PublicMemory:   publicMemory,
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

func (runner *Runner) GetAirPrivateInput(tracePath, memoryPath string) (AirPrivateInput, error) {
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
