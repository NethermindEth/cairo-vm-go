package builtins

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"

	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

const KeccakName = "keccak"
const cellsPerKeccak = 16
const inputCellsPerKeccak = 8

type Keccak struct {
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

	out := keccakF1600(dataU64)

	var output [200]byte
	for i := 0; i < 25; i++ {
		v := binary.LittleEndian.AppendUint64(nil, out[i])
		copy(output[i*8:i*8+8], v)
	}

	for i := 0; i < inputCellsPerKeccak; i++ {
		var bytes [32]byte
		copy(bytes[:], output[i*25:i*25+25])
		//Only 200 bits so always fits, no need to check error
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

const (
	KECCAK_FULL_RATE_IN_U64S  = 17
	KECCAK_FULL_RATE_IN_BYTES = 136
	BYTES_IN_U64_WORD         = 8
)

// U128Split splits a uint128 (represented as a uint256.Int) into two uint64s.
func U128Split(input *uint256.Int) (high, low uint64) {
	low = input.Uint64()
	input.Rsh(input, 64)
	high = input.Uint64()
	return
}

// Keccak256 computes the Keccak-256 hash of the input data.
func Keccak256(data []byte) ([]byte, error) {
	hasher := sha3.NewLegacyKeccak256()
	if _, err := hasher.Write(data); err != nil {
		return nil, err // Return the error to the caller.
	}
	return hasher.Sum(nil), nil // Return the hash and a nil error.
}

// ConvertToByteData converts a slice of uint64 to a slice of byte.
func ConvertToByteData(input []uint64) []byte {
	byteData := make([]byte, len(input)*8) // 8 bytes per uint64
	for i, word := range input {
		binary.LittleEndian.PutUint64(byteData[i*8:], word)
	}

	// To verify from web https://emn178.github.io/online-tools/keccak_256.html
	fmt.Printf("ByteData result after padding, before Keccak256 hashing: \n%x\n", byteData)

	return byteData
}

func CairoKeccak(input []uint64, lastInputWord uint64, lastInputNumBytes int) ([]byte, error) {
	//Add Padding to the input
	paddedInput, err := AddPadding(input, lastInputWord, lastInputNumBytes)
	if err != nil {
		return nil, err
	}
	//convert from uint64 to byte slice and then apply Keccak
	return Keccak256(ConvertToByteData(paddedInput))
}

func AddPadding(input []uint64, lastInputWord uint64, lastInputNumBytes int) ([]uint64, error) {
	fmt.Printf("input before AddPadding conversion in addpadding: %x\n ", input)
	wordsDivisor := KECCAK_FULL_RATE_IN_U64S
	lastBlockNumFullWords := len(input) % wordsDivisor

	var firstWordToAppend, firstPaddingBytePart, r uint64
	if lastInputNumBytes == 0 {
		firstWordToAppend = 1
	} else {
		switch lastInputNumBytes {
		case 1:
			firstPaddingBytePart = 0x100
		case 2:
			firstPaddingBytePart = 0x10000
		case 3:
			firstPaddingBytePart = 0x1000000
		case 4:
			firstPaddingBytePart = 0x100000000
		case 5:
			firstPaddingBytePart = 0x10000000000
		case 6:
			firstPaddingBytePart = 0x1000000000000
		case 7:
			firstPaddingBytePart = 0x100000000000000
		default:
			return nil, fmt.Errorf("keccak last input word >7b")
		}
		r = lastInputWord % firstPaddingBytePart
		firstWordToAppend = firstPaddingBytePart + r
	}

	//	// Debug print statements:
	fmt.Printf("firstPaddingBytePart: %x\n", firstPaddingBytePart)
	fmt.Printf("r: %x\n", r)
	fmt.Printf("firstWordToAppend: %x\n", firstWordToAppend)

	inputAfterPadding := append(input, firstWordToAppend)

	if lastBlockNumFullWords == KECCAK_FULL_RATE_IN_U64S-1 {
		inputAfterPadding = append(inputAfterPadding, 0x8000000000000000+firstWordToAppend)
		return inputAfterPadding, nil
	}

	//Debug when error
	fmt.Println("Before finalizePadding:", inputAfterPadding)
	finalizedPadding := finalizePadding(inputAfterPadding, uint32(KECCAK_FULL_RATE_IN_U64S-1-lastBlockNumFullWords))
	fmt.Println("After finalizePadding:", finalizedPadding)

	return finalizedPadding, nil
}

func finalizePadding(input []uint64, numPaddingWords uint32) []uint64 {
	for i := uint32(0); i < numPaddingWords; i++ {
		if i == numPaddingWords-1 {
			input = append(input, 0x8000000000000000)
		} else {
			input = append(input, 0)
		}
	}
	return input
}

//Little Endian way for Keccak256

// KeccakAddU256LE appends the low and high 64-bit words of a U256 value to a keccak input.
func KeccakAddU256LE(keccakInput []uint64, value *uint256.Int) []uint64 {
	valueCopy := new(uint256.Int).Set(value)
	// Split the "low" 128 bits into two uint64s and append them to keccakInput.
	low, high := U128Split(valueCopy)
	keccakInput = append(keccakInput, low, high)

	// Shift the value right by 128 bits to get the high 128 bits.
	valueCopy.Rsh(valueCopy, 128)

	// Split the "high" 128 bits into two uint64s and append them to keccakInput.
	low, high = U128Split(valueCopy)
	keccakInput = append(keccakInput, low, high)

	return keccakInput
}

// 1 initializes an array (u64)
// 2 converting input into array
// 3 adds padding to the array
// 4 put that array into keccak hashing

func KeccakU256sLEInputs(inputs []uint256.Int) ([]byte, error) {
	var keccakInput []uint64
	for _, input := range inputs {
		keccakInput = KeccakAddU256LE(keccakInput, &input)
	}
	return CairoKeccak(keccakInput, 0, 0)
}

//Big Endian way codes

func ReverseBytes128(value *uint256.Int) *uint256.Int {
	byteSlice := value.Bytes()
	reversedSlice := make([]byte, len(byteSlice))

	for i, byteValue := range byteSlice {
		reversedSlice[len(byteSlice)-1-i] = byteValue
	}

	reversed := uint256.NewInt(0)
	reversed.SetBytes(reversedSlice)

	return reversed
}

func KeccakAddU256BE(keccakInput []uint64, value *uint256.Int) []uint64 {
	valueCopy := new(uint256.Int).Set(value)
	valueHigh, valueLow := new(uint256.Int), new(uint256.Int)
	valueHigh.Rsh(valueCopy, 128)
	maxU128 := utils.Uint256Max128
	valueLow.And(valueCopy, &maxU128)

	reversedHigh := ReverseBytes128(valueHigh)
	reversedLow := ReverseBytes128(valueLow)

	keccakInput = KeccakAddU256LE(keccakInput, reversedHigh)
	keccakInput = KeccakAddU256LE(keccakInput, reversedLow)
	return keccakInput
}

func KeccakU256sBEInputs(inputs []uint256.Int) ([]byte, error) {
	var keccakInput []uint64
	for _, value := range inputs {
		keccakInput = KeccakAddU256BE(keccakInput, &value)
	}
	return CairoKeccak(keccakInput, 0, 0)
}
