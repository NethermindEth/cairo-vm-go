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
