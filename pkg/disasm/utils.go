package disasm

import (
	"strconv"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func feltToInt64(felt *fp.Element) int64 {
	// This would not be correct: int64(felt.Uint64)
	// since signed values will reside in more than one 64-bit word.
	//
	// BigInt().Int64() would not work neither.
	//
	// String() handles signed values pretty well for our use-case.
	// Maybe there is another way to avoid the redundant String()+Parsing?

	s := felt.String()
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}
