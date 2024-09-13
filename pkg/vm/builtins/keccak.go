package builtins

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

//
// todo(jkktom) find a way to let users know how to implement or make use of Keccak
// Below includes direct implementation of Keccak (Keccak256).
// Note that it's composed of two sections, implementing F1600 (Keccak F1600) as builtins,
// and also as Keccak Library for additional use (Keccak related hints or directly using it)
// It's useful to give users options to use Keccak just as Rust VM does it with it's keccak.cairo as library.
//

const (
	KeccakName                  = "keccak"
	cellsPerKeccak              = 16
	inputCellsPerKeccak         = 8
	instancesPerComponentKeccak = 16
)

type Keccak struct {
	ratio uint64
	cache map[uint64]fp.Element
}

func (k *Keccak) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (k *Keccak) InferValue(segment *memory.Segment, offset uint64) error {
	value, ok := k.cache[offset]
	if ok {
		mv := memory.MemoryValueFromFieldElement(&value)
		return segment.Write(offset, &mv)
	}
	hashIndex := offset % cellsPerKeccak
	if hashIndex < inputCellsPerKeccak {
		return errors.New("cannot infer value")
	}

	startOffset := offset - hashIndex
	var data [200]byte
	for i := uint64(0); i < inputCellsPerKeccak; i++ {
		value := segment.Peek(startOffset + i)
		if !value.Known() {
			return fmt.Errorf("cannot infer value: input value at offset %d is unknown", startOffset+i)
		}
		v, err := value.FieldElement()
		if err != nil {
			return fmt.Errorf("Keccak input has to be felt")
		}
		var out [32]byte
		fp.LittleEndian.PutElement(&out, *v)
		copy(data[i*25:i*25+25], out[:25]) //25*8 = 200bits
	}

	var dataU64 [25]uint64
	for i := 0; i < 25; i++ {
		dataU64[i] = binary.LittleEndian.Uint64(data[8*i : 8*i+8])
	}

	KeccakF1600(&dataU64)

	var output [200]byte
	for i := 0; i < 25; i++ {
		binary.LittleEndian.PutUint64(output[i*8:i*8+8], dataU64[i])
	}

	for i := 0; i < inputCellsPerKeccak; i++ {
		var bytes [32]byte
		copy(bytes[:], output[i*25:i*25+25])
		//This is 25*8 bits which is smaller than max felt 252 bits so no need to check the error
		v, _ := fp.LittleEndian.Element(&bytes)
		k.cache[startOffset+inputCellsPerKeccak+uint64(i)] = v
	}
	value = k.cache[offset]
	mv := memory.MemoryValueFromFieldElement(&value)
	return segment.Write(offset, &mv)
}

func (k *Keccak) String() string {
	return KeccakName
}

func (k *Keccak) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	return getBuiltinAllocatedSize(segmentUsedSize, vmCurrentStep, k.ratio, inputCellsPerKeccak, instancesPerComponentKeccak, cellsPerKeccak)
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

func (k *Keccak) GetAirPrivateInput(keccakSegment *memory.Segment) []AirPrivateBuiltinKeccak {
	valueMapping := make(map[int]AirPrivateBuiltinKeccak)
	for index, value := range keccakSegment.Data {
		if !value.Known() {
			continue
		}
		idx, stateIndex := index/cellsPerKeccak, index%cellsPerKeccak
		if stateIndex >= inputCellsPerKeccak {
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

	values := make([]AirPrivateBuiltinKeccak, 0)

	sortedIndexes := make([]int, 0, len(valueMapping))
	for index := range valueMapping {
		sortedIndexes = append(sortedIndexes, index)
	}
	sort.Ints(sortedIndexes)
	for _, index := range sortedIndexes {
		value := valueMapping[index]
		values = append(values, value)
	}
	return values
}
