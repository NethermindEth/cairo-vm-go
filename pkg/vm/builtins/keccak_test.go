package builtins

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeccak256(t *testing.T) {
	// Input data: array of bytes from 1 to 17, similar to the input in the Rust test.
	input := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}
	hexStr := hex.EncodeToString(input)
	fmt.Println("the hex Str is: ", hexStr)
	//0102030405060708090a0b0c0d0e0f1011

	// Known correct Keccak-256 hash value for the input data, obtained from an online calculator. https://emn178.github.io/online-tools/keccak_256.html
	expectedHash := "57e46e893805ca9503660cd6ef2b39e24750a502ed7c71270f214dcb114b7113"
	expectedHashByteData, err := hex.DecodeString(expectedHash)
	fmt.Println("the expectedHash to byte is: ", expectedHashByteData, err)

	// Call the Keccak256 function.
	hash, err := Keccak256(input)
	require.NoError(t, err) // Ensure no error is returned.

	// Convert the obtained hash to a hex string.
	hashHex := hex.EncodeToString(hash)
	fmt.Println("the hash is: ", hash)

	// Compare the obtained hash with the expected hash.
	assert.Equal(t, expectedHash, hashHex, "Expected %s, got %s", expectedHash, hashHex)

	// for keccak result : https://emn178.github.io/online-tools/keccak_256.html?input_type=hex&input=0102030405060708090a0b0c0d0e0f1011
	// which is
	//57e46e893805ca9503660cd6ef2b39e24750a502ed7c71270f214dcb114b7113
	// for sha3 result : https://emn178.github.io/online-tools/sha3_256.html?input_type=hex&input=0102030405060708090a0b0c0d0e0f1011
	// which is
	//5ea232134fe405605b468996dd5d77e7cdf6b63be625826d4f27c302ab17d65a
	//Rust output is : d2eb808dfba4703c528d145dfe6571afec687be9c50d2218388da73622e8fdd5
	//D2EB808DFBA4703C528D145DFE6571AFEC687BE9C50D2218388DA73622E8FDD5

}

func TestU128Split(t *testing.T) {
	tests := []struct {
		input        string
		expectedHigh uint64
		expectedLow  uint64
	}{
		{
			input:        "00000001000000020000000300000004",
			expectedHigh: 0x0000000100000002,
			expectedLow:  0x0000000300000004,
		},
	}

	for _, test := range tests {
		trimmedInput := strings.TrimLeft(test.input, "0")
		if len(trimmedInput) == 0 {
			trimmedInput = "0" // Ensure at least one zero remains if all digits are zero.
		}
		input, err := uint256.FromHex("0x" + trimmedInput)
		if err != nil {
			t.Fatalf("failed to parse input %s: %v", test.input, err)
		}
		high, low := U128Split(input)
		if high != test.expectedHigh || low != test.expectedLow {
			t.Errorf("expected (%d, %d), got (%d, %d) for input %s", test.expectedHigh, test.expectedLow, high, low, test.input)
		}
	}
}

func TestAddPaddingPlainCase(t *testing.T) {
	// Define your test cases as a table
	testCases := []struct {
		input             []uint64
		lastInputWord     uint64
		lastInputNumBytes int
		expected          []uint64
	}{
		{
			input:             []uint64{},
			lastInputWord:     0,
			lastInputNumBytes: 0,
			expected:          []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x8000000000000000},
		},
		{
			input: []uint64{
				0x0000000000000001,
				0x0000000000000000,
				0x0000000000000000,
				0x0000000000000000,
			},
			lastInputWord:     0,
			lastInputNumBytes: 0,
			expected: []uint64{
				0x0000000000000001, // 1st
				0x0000000000000000, // 2nd
				0x0000000000000000, // 3rd
				0x0000000000000000, // 4th
				0x0000000000000001, // 5th
				0x0000000000000000, // 6th
				0x0000000000000000, // 7th
				0x0000000000000000, // 8th
				0x0000000000000000, // 9th
				0x0000000000000000, // 10th
				0x0000000000000000, // 11th
				0x0000000000000000, // 12th
				0x0000000000000000, // 13th
				0x0000000000000000, // 14th
				0x0000000000000000, // 15th
				0x0000000000000000, // 16th
				0x8000000000000000, // 17th
			},
		},
		{
			input:             []uint64{0x1234567890abcdef},
			lastInputWord:     0,
			lastInputNumBytes: 0,
			expected:          []uint64{0x1234567890abcdef, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x8000000000000000},
		},
	}

	for _, testCase := range testCases {
		result, err := AddPadding(testCase.input, testCase.lastInputWord, testCase.lastInputNumBytes)
		if assert.NoError(t, err) {
			assert.Equal(t, testCase.expected, result)
		}
	}
}

func TestAddPaddingOperandCase(t *testing.T) {
	// Define your test cases as a table
	testCases := []struct {
		input             []uint64
		lastInputWord     uint64
		lastInputNumBytes int
		expected          []uint64
	}{
		{
			input:             []uint64{0x1234567890abcdef},
			lastInputWord:     0xabcdef,
			lastInputNumBytes: 3,
			expected:          []uint64{0x1234567890abcdef, 0x1000000 + 0xabcdef, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x8000000000000000},
		},
	}

	for _, testCase := range testCases {
		result, err := AddPadding(testCase.input, testCase.lastInputWord, testCase.lastInputNumBytes)
		if assert.NoError(t, err) {
			assert.Equal(t, testCase.expected, result)
		}
	}
}

func TestCairoKeccakU64(t *testing.T) {
	tests := []struct {
		name              string
		input             []uint64
		lastInputWord     uint64
		lastInputNumBytes int
		expectedLow       string
		expectedHigh      string
	}{
		{
			name: "Test case 1 from Rust",
			input: []uint64{
				0x0000000000000001,
				0x0000000000000000,
				0x0000000000000000,
				0x0000000000000000,
			},
			lastInputWord:     0,
			lastInputNumBytes: 0,
			expectedLow:       "0x587f7cc3722e9654ea3963d5fe8c0748",
			expectedHigh:      "0xa5963aa610cb75ba273817bce5f8c48f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CairoKeccak(tt.input, tt.lastInputWord, tt.lastInputNumBytes)
			require.NoError(t, err)

			// Assuming res is a []byte that represents a 256-bit number,
			// split it into two 128-bit numbers represented as strings.
			low := fmt.Sprintf("%x", res[:16])
			high := fmt.Sprintf("%x", res[16:])
			require.Equal(t, tt.expectedLow, low)
			require.Equal(t, tt.expectedHigh, high)
		})
	}
}

// [
// 	1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0   // 1
// 	0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0   // 2
// 	0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0   // 3
// 	0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0   // 4
// 	0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0   // 5
// 	0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0   // 6
// 	0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0   // 7
// 	0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 128 // 8
// ]
