package utils

import (
	"testing"
)

func TestRightRot(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint32
		n        uint32
		expected uint32
	}{
		{
			name:     "Rotate 0 bits",
			value:    0x12345678,
			n:        0,
			expected: 0x12345678,
		},
		{
			name:     "Rotate 1 bit",
			value:    0x12345678,
			n:        1,
			expected: 0x091A2B3C,
		},
		{
			name:     "Rotate 4 bits",
			value:    0x12345678,
			n:        4,
			expected: 0x81234567,
		},
		{
			name:     "Rotate 8 bits",
			value:    0x12345678,
			n:        8,
			expected: 0x78123456,
		},
		{
			name:     "Rotate 16 bits",
			value:    0x12345678,
			n:        16,
			expected: 0x56781234,
		},
		{
			name:     "Rotate 31 bits",
			value:    0x12345678,
			n:        31,
			expected: 0x2468ACF0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := rightRot(tc.value, tc.n)
			if result != tc.expected {
				t.Errorf("Expected %08X, got %08X", tc.expected, result)
			}
		})
	}
}

// Result from CairoVM's blake2s implementation in Rust by Lambdaclass
// https://play.rust-lang.org/?version=stable&mode=debug&edition=2021&gist=987596d755c09f8803a097d29594f92c

func TestMix(t *testing.T) {
	// Test case
	a := uint32(0x11223344)
	b := uint32(0x55667788)
	c := uint32(0x99AABBCC)
	d := uint32(0xDDEEFF00)
	m0 := uint32(0x12345678)
	m1 := uint32(0x9ABCDEF0)

	expectedA := uint32(0x7CF608C5)
	expectedB := uint32(0xBA8E1C76)
	expectedC := uint32(0x2E7213CC)
	expectedD := uint32(0x9682B2AD)

	resultA, resultB, resultC, resultD := mix(a, b, c, d, m0, m1)

	if resultA != expectedA {
		t.Errorf("Expected a = %08X, got %08X", expectedA, resultA)
	}
	if resultB != expectedB {
		t.Errorf("Expected b = %08X, got %08X", expectedB, resultB)
	}
	if resultC != expectedC {
		t.Errorf("Expected c = %08X, got %08X", expectedC, resultC)
	}
	if resultD != expectedD {
		t.Errorf("Expected d = %08X, got %08X", expectedD, resultD)
	}
}

func TestBlake2sComppress(t *testing.T) {
	testCases := []struct {
		name           string
		message        []uint32
		h              [8]uint32
		compressParams [4]uint32
		expected       [8]uint32
	}{
		{
			name:           "Test case 1",
			message:        []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			h:              [8]uint32{1795745351, 3144134277, 1013904242, 2773480762, 1359893119, 2600822924, 528734635, 1541459225},
			compressParams: [4]uint32{2, 0, 4294967295, 0},
			expected:       [8]uint32{412110711, 3234706100, 3894970767, 982912411, 937789635, 742982576, 3942558313, 1407547065},
		},
		{
			name:           "Test case 2",
			message:        []uint32{1819043144, 1870078063, 6581362, 274628678, 715791845, 175498643, 871587583, 635963558, 557369694, 1576875962, 215769785, 0, 0, 0, 0, 0},
			h:              [8]uint32{1795745351, 3144134277, 1013904242, 2773480762, 1359893119, 2600822924, 528734635, 1541459225},
			compressParams: [4]uint32{44, 0, 4294967295, 0},
			expected:       [8]uint32{3251785223, 1946079609, 2665255093, 3508191500, 3630835628, 3067307230, 3623370123, 656151356},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Blake2sCompress(tc.message, tc.h, tc.compressParams[0], tc.compressParams[1], tc.compressParams[2], tc.compressParams[3])
			for i := 0; i < 8; i++ {
				if result[i] != tc.expected[i] {
					t.Errorf("Expected %08X, got %08X", tc.expected[i], result[i])
				}
			}
		})
	}
}
