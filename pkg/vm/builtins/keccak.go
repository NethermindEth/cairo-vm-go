package builtins

import (
	"encoding/binary"
	"errors"
	"fmt"

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

const KeccakName = "keccak"
const cellsPerKeccak = 16
const inputCellsPerKeccak = 8

// TODO: This is from layout small, those values should be dynamically loaded from given layout
// const ratioKeccak = 8
// const instancesPerComponentKeccak = 5

type Keccak struct {
	ratioKeccak                 uint64
	instancesPerComponentKeccak uint64
}

func (k *Keccak) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (k *Keccak) InferValue(segment *memory.Segment, offset uint64) error {
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

	keccakF1600(&dataU64)

	var output [200]byte
	for i := 0; i < 25; i++ {
		binary.LittleEndian.PutUint64(output[i*8:i*8+8], dataU64[i])
	}

	for i := 0; i < inputCellsPerKeccak; i++ {
		var bytes [32]byte
		copy(bytes[:], output[i*25:i*25+25])
		//This is 25*8 bits which is smaller than max felt 252 bits so no need to check the error
		v, _ := fp.LittleEndian.Element(&bytes)
		mv := memory.MemoryValueFromFieldElement(&v)
		err := segment.Write(startOffset+inputCellsPerKeccak+uint64(i), &mv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *Keccak) String() string {
	return KeccakName
}

func (k *Keccak) GetAllocatedSize(segmentUsedSize uint64, vmCurrentStep uint64) (uint64, error) {
	allocatedInstances, err := GetAllocatedInstances(k.ratioKeccak, inputCellsPerKeccak, segmentUsedSize, k.instancesPerComponentKeccak, vmCurrentStep)
	if err != nil {
		return 0, err
	}
	return allocatedInstances * cellsPerKeccak, nil
}
