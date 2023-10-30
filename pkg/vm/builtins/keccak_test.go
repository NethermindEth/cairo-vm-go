package builtins

import (
	"bytes"
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

	//To verify result from the web
	// for keccak result : https://emn178.github.io/online-tools/keccak_256.html?input_type=hex&input=0102030405060708090a0b0c0d0e0f1011
	// which is
	//57e46e893805ca9503660cd6ef2b39e24750a502ed7c71270f214dcb114b7113
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

func TestCairoKeccak(t *testing.T) {
	tests := []struct {
		name              string
		input             []uint64
		lastInputWord     uint64
		lastInputNumBytes int
		expectedHash      string
	}{
		{
			name:              "Test case 1",
			input:             []uint64{1, 0, 0, 0},
			lastInputWord:     0,
			lastInputNumBytes: 0,
			expectedHash:      "a80c226a0612d2578acff9caeb00583cb8002ccfb5f442d47ec6838f35d72b2c",
			//https://emn178.github.io/online-tools/keccak_256.html?input_type=hex&input=01000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := CairoKeccak(tt.input, tt.lastInputWord, tt.lastInputNumBytes)
			require.NoError(t, err)

			hashHex := hex.EncodeToString(res)
			fmt.Println("the hash is: ", hashHex)

			// Compare the obtained hash with the expected hash.
			assert.Equal(t, tt.expectedHash, hashHex, "Expected %s, got %s", tt.expectedHash, hashHex)
		})
	}
}

func TestKeccakU256sLEInputs(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected string
	}{
		{
			input:    uint64(1),
			expected: "173af0af902b536af63b93f3f6dc2d7369e5b06780ffe7ce5892797e2bb1d23b",
		},
		{
			input:    uint256.NewInt(123456789),
			expected: "350187ae11fa6a4e2feead1a234e3d700888faec6db7b99cfeb99d08cd21f383",
		},
		{
			input:    uint256.NewInt(1).Lsh(uint256.NewInt(1), 256).Sub(uint256.NewInt(1), uint256.NewInt(1)),
			expected: "b6b62fa9dcc719ee570994c986575cdc696fe70949d66ef63248ab4302563bcd",
		},
		{
			input:    createUint256IntFromBytes([]byte{1, 2, 3, 4}), // Create a uint256.Int from byte slice
			expected: "0b162d3055df4f1f6f28b54c2fca57cb224c2b2f00e266ea305213d62b0447c6",
		},
		{
			input:    StringToUint256("Hello world"), // convert string to uint256
			expected: "7eb2e382db7d2a31f2400a62c39e4ec5486d76734442402731987b8c99280fea",
		},
	}

	for _, c := range cases {
		convInput := ConvertToUint256(c.input)
		if convInput == nil {
			t.Errorf("Conversion to uint256.Int failed for input: %v", c.input)
			continue
		}
		hash, err := KeccakU256sLEInputs([]uint256.Int{*convInput})
		if err != nil {
			t.Errorf("KeccakU256sLEInputs returned an error: %v", err)
		}

		// Convert the expected hex string to bytes for comparison
		expectedBytes, err := hex.DecodeString(c.expected)
		if err != nil {
			t.Errorf("Failed to decode expected hex string: %v", err)
		}

		// Compare the hash with the expected bytes
		if len(hash) != len(expectedBytes) || !bytes.Equal(hash, expectedBytes) {
			t.Errorf("KeccakU256sLEInputs = %x, want %x", hash, expectedBytes)
		}
	}
}

// Helper Function for converting various inputs to uint256
func ConvertToUint256(input interface{}) *uint256.Int {
	switch v := input.(type) {
	case string:
		// Convert string to uint256.Int
		return StringToUint256(v)
	case []byte:
		// Convert byte slice to uint256.Int
		return createUint256IntFromBytes(v)
	case uint64:
		// Convert uint64 to uint256.Int
		return uint256.NewInt(v)
	case *uint256.Int:
		// Directly use the *uint256.Int
		return v
	default:
		return nil // or handle error
	}
}

// This function converts a string to a uint256.Int by taking the byte representation
// of the string and using it as the least significant bytes of the uint256.Int.
// This is for testing purposes.
func StringToUint256(s string) *uint256.Int {
	bytes := []byte(s)
	intVal := uint256.NewInt(0)

	// This loop adds each byte into the uint256.Int. It's a simple conversion and
	// doesn't handle strings longer than 32 bytes (256 bits).
	for _, b := range bytes {
		intVal.Lsh(intVal, 8)                        // Shift left by 8 bits to make room for the next byte
		intVal.Or(intVal, uint256.NewInt(uint64(b))) // Add the new byte
	}

	return intVal
}

// Helper function to convert a byte slice into a uint256.Int
func createUint256IntFromBytes(b []byte) *uint256.Int {
	// Pad the byte slice to 32 bytes if necessary
	for len(b) < 32 {
		b = append(b, 0)
	}
	value := new(uint256.Int)
	value.SetBytes(b)
	return value
}

// BE (Big Endian)

func TestKeccakU256sBEInputs(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected string
	}{
		{
			input:    uint64(1),
			expected: "bc142d192087ce36660c866a686ee4b129ad5591ecb90152c2fc373793a2dcb5",
		},
		{
			input:    uint256.NewInt(123456789),
			expected: "b29ac0dd0032cc7e4f979c07c5c2b631eb5fb6438cda0b2e09990f5b5264e9be",
		},
		{
			input:    uint256.NewInt(1).Lsh(uint256.NewInt(1), 256).Sub(uint256.NewInt(1), uint256.NewInt(1)),
			expected: "97a2173291d40ef311333110eb76dd78bd4b38ae6d510eded63831d76c069514",
		},
		{
			input:    createUint256IntFromBytes([]byte{1, 2, 3, 4}), // Create a uint256.Int from byte slice
			expected: "6bcb98eb7d2cd01c55f355cf4ec727573517c1101d35eb0445f741029ec7694b",
		},
		{
			input:    StringToUint256("Hello world"), // convert string to uint256
			expected: "7e98083eaca37e63048eb2863b36d14ffe3a5a331346e9cd968329c76dfac50d",
		},
	}

	for _, c := range cases {
		convInput := ConvertToUint256(c.input)
		if convInput == nil {
			t.Errorf("Conversion to uint256.Int failed for input: %v", c.input)
			continue
		}
		hash, err := KeccakU256sBEInputs([]uint256.Int{*convInput})
		if err != nil {
			t.Errorf("KeccakU256sLEInputs returned an error: %v", err)
		}

		// Convert the expected hex string to bytes for comparison
		expectedBytes, err := hex.DecodeString(c.expected)
		if err != nil {
			t.Errorf("Failed to decode expected hex string: %v", err)
		}

		// Compare the hash with the expected bytes
		if len(hash) != len(expectedBytes) || !bytes.Equal(hash, expectedBytes) {
			t.Errorf("KeccakU256sLEInputs = %x, want %x", hash, expectedBytes)
		}
	}
}
