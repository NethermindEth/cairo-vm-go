package utils

import (
	"math/big"
	"testing"
)

func TestDivMod(t *testing.T) {
	// https://github.com/starkware-libs/cairo-lang/blob/efa9648f57568aad8f8a13fbf027d2de7c63c2c0/src/starkware/python/math_utils_test.py#L108
	tests := []struct {
		name           string
		n, m, p        *big.Int
		expected       *big.Int
		expectedErrMsg string
	}{
		{
			name:     "Basic case",
			n:        big.NewInt(2),
			m:        big.NewInt(3),
			p:        big.NewInt(5),
			expected: big.NewInt(4),
		},
		{
			name:           "Error case",
			n:              big.NewInt(8),
			m:              big.NewInt(10),
			p:              big.NewInt(5),
			expectedErrMsg: "no solution exists (gcd(m, p) != 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Divmod(tt.n, tt.m, tt.p)

			if err != nil {
				if err.Error() != tt.expectedErrMsg {
					t.Errorf("got error: %v, want: %v", err, tt.expectedErrMsg)
				}
				return
			}

			if actual.Cmp(tt.expected) != 0 {
				t.Errorf("got quotient: %v, want: %v", actual, tt.expected)
			}
		})
	}
}

func TestIgcdex(t *testing.T) {
	// https://github.com/sympy/sympy/blob/e7fb2714f17b30b83e424448aad0da9e94a4b577/sympy/core/tests/test_numbers.py#L278
	tests := []struct {
		name                            string
		a, b                            *big.Int
		expectedX, expectedY, expectedG *big.Int
	}{
		{
			name:      "Case 1",
			a:         big.NewInt(2),
			b:         big.NewInt(3),
			expectedX: big.NewInt(-1),
			expectedY: big.NewInt(1),
			expectedG: big.NewInt(1),
		},
		{
			name:      "Case 2",
			a:         big.NewInt(10),
			b:         big.NewInt(12),
			expectedX: big.NewInt(-1),
			expectedY: big.NewInt(1),
			expectedG: big.NewInt(2),
		},
		{
			name:      "Case 3",
			a:         big.NewInt(100),
			b:         big.NewInt(2004),
			expectedX: big.NewInt(-20),
			expectedY: big.NewInt(1),
			expectedG: big.NewInt(4),
		},
		{
			name:      "Case 4",
			a:         big.NewInt(0),
			b:         big.NewInt(0),
			expectedX: big.NewInt(0),
			expectedY: big.NewInt(1),
			expectedG: big.NewInt(0),
		},
		{
			name:      "Case 5",
			a:         big.NewInt(1),
			b:         big.NewInt(0),
			expectedX: big.NewInt(1),
			expectedY: big.NewInt(0),
			expectedG: big.NewInt(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualX, actualY, actualG := Igcdex(tt.a, tt.b)

			if actualX.Cmp(tt.expectedX) != 0 {
				t.Errorf("got x: %v, want: %v", actualX, tt.expectedX)
			}
			if actualY.Cmp(tt.expectedY) != 0 {
				t.Errorf("got x: %v, want: %v", actualY, tt.expectedY)
			}
			if actualG.Cmp(tt.expectedG) != 0 {
				t.Errorf("got x: %v, want: %v", actualG, tt.expectedG)
			}
		})
	}
}
