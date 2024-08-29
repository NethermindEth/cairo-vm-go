package zero

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
)

func (runner *ZeroRunner) GetAirPublicInput() (AirPublicInput, error) {
	rcMin, rcMax := runner.getPermRangeCheckLimits()
	return AirPublicInput{
		Layout:        runner.layout.Name,
		RcMin:         rcMin,
		RcMax:         rcMax,
		NSteps:        len(runner.vm.Trace),
		DynamicParams: nil,
		// TODO: yet to be implemented
		MemorySegments: make(map[string]AirMemorySegmentEntry),
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
	BeginAddr int `json:"begin_addr"`
	StopPtr   int `json:"stop_ptr"`
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
			case "range_check":
				{
					builtinValues := make([]AirPrivateBuiltinRangeCheck, 0)
					for index, value := range builtinSegment.Data {
						if !value.Known() {
							continue
						}
						valueBig := big.Int{}
						value.Felt.BigInt(&valueBig)
						valueHex := fmt.Sprintf("0x%x", &valueBig)
						builtinValues = append(builtinValues, AirPrivateBuiltinRangeCheck{Index: index, Value: valueHex})
					}
					airPrivateInput.RangeCheck = builtinValues
				}
			case "bitwise":
				{
					valueMapping := make(map[int]AirPrivateBuiltinBitwise)
					for index, value := range builtinSegment.Data {
						if !value.Known() {
							continue
						}
						idx, typ := index/builtins.CellsPerBitwise, index%builtins.CellsPerBitwise
						if typ >= 2 {
							continue
						}

						builtinValue, exists := valueMapping[idx]
						if !exists {
							builtinValue = AirPrivateBuiltinBitwise{Index: idx}
						}

						valueBig := big.Int{}
						value.Felt.BigInt(&valueBig)
						valueHex := fmt.Sprintf("0x%x", &valueBig)
						if typ == 0 {
							builtinValue.X = valueHex
						} else {
							builtinValue.Y = valueHex
						}
						valueMapping[idx] = builtinValue
					}

					builtinValues := make([]AirPrivateBuiltinBitwise, 0)

					sortedIndexes := make([]int, 0, len(valueMapping))
					for index := range valueMapping {
						sortedIndexes = append(sortedIndexes, index)
					}
					sort.Ints(sortedIndexes)
					for _, index := range sortedIndexes {
						builtinValue := valueMapping[index]
						builtinValues = append(builtinValues, builtinValue)
					}

					airPrivateInput.Bitwise = builtinValues
				}
			case "poseidon":
				{
					valueMapping := make(map[int]AirPrivateBuiltinPoseidon)
					for index, value := range builtinSegment.Data {
						if !value.Known() {
							continue
						}
						idx, stateIndex := index/builtins.CellsPerPoseidon, index%builtins.CellsPerPoseidon
						if stateIndex >= builtins.InputCellsPerPoseidon {
							continue
						}

						builtinValue, exists := valueMapping[idx]
						if !exists {
							builtinValue = AirPrivateBuiltinPoseidon{Index: idx}
						}

						valueBig := big.Int{}
						value.Felt.BigInt(&valueBig)
						valueHex := fmt.Sprintf("0x%x", &valueBig)
						if stateIndex == 0 {
							builtinValue.InputS0 = valueHex
						} else if stateIndex == 1 {
							builtinValue.InputS1 = valueHex
						} else if stateIndex == 2 {
							builtinValue.InputS2 = valueHex
						}
						valueMapping[idx] = builtinValue
					}

					builtinValues := make([]AirPrivateBuiltinPoseidon, 0)

					sortedIndexes := make([]int, 0, len(valueMapping))
					for index := range valueMapping {
						sortedIndexes = append(sortedIndexes, index)
					}
					sort.Ints(sortedIndexes)
					for _, index := range sortedIndexes {
						builtinValue := valueMapping[index]
						builtinValues = append(builtinValues, builtinValue)
					}

					airPrivateInput.Poseidon = builtinValues
				}
			case "pedersen":
				{
					valueMapping := make(map[int]AirPrivateBuiltinPedersen)
					for index, value := range builtinSegment.Data {
						if !value.Known() {
							continue
						}
						idx, typ := index/builtins.CellsPerPedersen, index%builtins.CellsPerPedersen
						if typ == 2 {
							continue
						}

						builtinValue, exists := valueMapping[idx]
						if !exists {
							builtinValue = AirPrivateBuiltinPedersen{Index: idx}
						}

						valueBig := big.Int{}
						value.Felt.BigInt(&valueBig)
						valueHex := fmt.Sprintf("0x%x", &valueBig)
						if typ == 0 {
							builtinValue.X = valueHex
						} else {
							builtinValue.Y = valueHex
						}
						valueMapping[idx] = builtinValue
					}

					builtinValues := make([]AirPrivateBuiltinPedersen, 0)

					sortedIndexes := make([]int, 0, len(valueMapping))
					for index := range valueMapping {
						sortedIndexes = append(sortedIndexes, index)
					}
					sort.Ints(sortedIndexes)
					for _, index := range sortedIndexes {
						builtinValue := valueMapping[index]
						builtinValues = append(builtinValues, builtinValue)
					}

					airPrivateInput.Pedersen = builtinValues
				}
			case "ec_op":
				{
					valueMapping := make(map[int]AirPrivateBuiltinEcOp)
					for index, value := range builtinSegment.Data {
						if !value.Known() {
							continue
						}
						idx, typ := index/builtins.CellsPerEcOp, index%builtins.CellsPerEcOp
						if typ >= builtins.InputCellsPerEcOp {
							continue
						}

						builtinValue, exists := valueMapping[idx]
						if !exists {
							builtinValue = AirPrivateBuiltinEcOp{Index: idx}
						}

						valueBig := big.Int{}
						value.Felt.BigInt(&valueBig)
						valueHex := fmt.Sprintf("0x%x", &valueBig)
						if typ == 0 {
							builtinValue.PX = valueHex
						} else if typ == 1 {
							builtinValue.PY = valueHex
						} else if typ == 2 {
							builtinValue.QX = valueHex
						} else if typ == 3 {
							builtinValue.QY = valueHex
						} else if typ == 4 {
							builtinValue.M = valueHex
						}
						valueMapping[idx] = builtinValue
					}

					builtinValues := make([]AirPrivateBuiltinEcOp, 0)

					sortedIndexes := make([]int, 0, len(valueMapping))
					for index := range valueMapping {
						sortedIndexes = append(sortedIndexes, index)
					}
					sort.Ints(sortedIndexes)
					for _, index := range sortedIndexes {
						builtinValue := valueMapping[index]
						builtinValues = append(builtinValues, builtinValue)
					}

					airPrivateInput.EcOp = builtinValues
				}
			case "keccak":
				{
					valueMapping := make(map[int]AirPrivateBuiltinKeccak)
					for index, value := range builtinSegment.Data {
						if !value.Known() {
							continue
						}
						idx, stateIndex := index/builtins.CellsPerKeccak, index%builtins.CellsPerKeccak
						if stateIndex >= builtins.InputCellsPerKeccak {
							continue
						}

						builtinValue, exists := valueMapping[idx]
						if !exists {
							builtinValue = AirPrivateBuiltinKeccak{Index: idx}
						}

						valueBig := big.Int{}
						value.Felt.BigInt(&valueBig)
						valueHex := fmt.Sprintf("0x%x", &valueBig)
						if stateIndex == 0 {
							builtinValue.InputS0 = valueHex
						} else if stateIndex == 1 {
							builtinValue.InputS1 = valueHex
						} else if stateIndex == 2 {
							builtinValue.InputS2 = valueHex
						} else if stateIndex == 3 {
							builtinValue.InputS3 = valueHex
						} else if stateIndex == 4 {
							builtinValue.InputS4 = valueHex
						} else if stateIndex == 5 {
							builtinValue.InputS5 = valueHex
						} else if stateIndex == 6 {
							builtinValue.InputS6 = valueHex
						} else if stateIndex == 7 {
							builtinValue.InputS7 = valueHex
						}
						valueMapping[idx] = builtinValue
					}

					builtinValues := make([]AirPrivateBuiltinKeccak, 0)

					sortedIndexes := make([]int, 0, len(valueMapping))
					for index := range valueMapping {
						sortedIndexes = append(sortedIndexes, index)
					}
					sort.Ints(sortedIndexes)
					for _, index := range sortedIndexes {
						builtinValue := valueMapping[index]
						builtinValues = append(builtinValues, builtinValue)
					}

					airPrivateInput.Keccak = builtinValues
				}
			case "ecdsa":
				{
					ecdsaRunner, ok := bRunner.Runner.(*builtins.ECDSA)
					if !ok {
						return AirPrivateInput{}, fmt.Errorf("expected ECDSARunner")
					}

					builtinValues := make([]AirPrivateBuiltinECDSA, 0)
					for addrOffset, signature := range ecdsaRunner.Signatures {
						idx := addrOffset / builtins.CellsPerECDSA
						pubKey, err := builtinSegment.Read(addrOffset)
						if err != nil {
							return AirPrivateInput{}, err
						}
						msg, err := builtinSegment.Read(addrOffset + 1)
						if err != nil {
							return AirPrivateInput{}, err
						}

						pubKeyBig := big.Int{}
						msgBig := big.Int{}
						pubKey.Felt.BigInt(&pubKeyBig)
						msg.Felt.BigInt(&msgBig)
						pubKeyHex := fmt.Sprintf("0x%x", &pubKeyBig)
						msgHex := fmt.Sprintf("0x%x", &msgBig)

						rBig := new(big.Int).SetBytes(signature.R[:])
						sBig := new(big.Int).SetBytes(signature.S[:])
						frModulusBig, _ := new(big.Int).SetString("3618502788666131213697322783095070105526743751716087489154079457884512865583", 10)
						wBig := new(big.Int).ModInverse(sBig, frModulusBig)
						signatureInput := AirPrivateBuiltinECDSASignatureInput{
							R: fmt.Sprintf("0x%x", rBig),
							W: fmt.Sprintf("0x%x", wBig),
						}

						builtinValues = append(builtinValues, AirPrivateBuiltinECDSA{Index: int(idx), PubKey: pubKeyHex, Msg: msgHex, SignatureInput: signatureInput})
					}
					airPrivateInput.Ecdsa = builtinValues
				}
			}
		}
	}

	return airPrivateInput, nil
}

type AirPrivateInput struct {
	TracePath  string                        `json:"trace_path"`
	MemoryPath string                        `json:"memory_path"`
	Pedersen   []AirPrivateBuiltinPedersen   `json:"pedersen"`
	RangeCheck []AirPrivateBuiltinRangeCheck `json:"range_check"`
	Ecdsa      []AirPrivateBuiltinECDSA      `json:"ecdsa"`
	Bitwise    []AirPrivateBuiltinBitwise    `json:"bitwise"`
	EcOp       []AirPrivateBuiltinEcOp       `json:"ec_op"`
	Keccak     []AirPrivateBuiltinKeccak     `json:"keccak"`
	Poseidon   []AirPrivateBuiltinPoseidon   `json:"poseidon"`
}

type AirPrivateBuiltinRangeCheck struct {
	Index int    `json:"index"`
	Value string `json:"value"`
}

type AirPrivateBuiltinBitwise struct {
	Index int    `json:"index"`
	X     string `json:"x"`
	Y     string `json:"y"`
}

type AirPrivateBuiltinPoseidon struct {
	Index   int    `json:"index"`
	InputS0 string `json:"input_s0"`
	InputS1 string `json:"input_s1"`
	InputS2 string `json:"input_s2"`
}

type AirPrivateBuiltinPedersen struct {
	Index int    `json:"index"`
	X     string `json:"x"`
	Y     string `json:"y"`
}

type AirPrivateBuiltinEcOp struct {
	Index int    `json:"index"`
	PX    string `json:"p_x"`
	PY    string `json:"p_y"`
	M     string `json:"m"`
	QX    string `json:"q_x"`
	QY    string `json:"q_y"`
}

type AirPrivateBuiltinKeccak struct {
	Index   int    `json:"index"`
	InputS0 string `json:"input_s0"`
	InputS1 string `json:"input_s1"`
	InputS2 string `json:"input_s2"`
	InputS3 string `json:"input_s3"`
	InputS4 string `json:"input_s4"`
	InputS5 string `json:"input_s5"`
	InputS6 string `json:"input_s6"`
	InputS7 string `json:"input_s7"`
}

type AirPrivateBuiltinECDSA struct {
	Index          int                                  `json:"index"`
	PubKey         string                               `json:"pubkey"`
	Msg            string                               `json:"msg"`
	SignatureInput AirPrivateBuiltinECDSASignatureInput `json:"signature_input"`
}

type AirPrivateBuiltinECDSASignatureInput struct {
	R string `json:"r"`
	W string `json:"w"`
}
