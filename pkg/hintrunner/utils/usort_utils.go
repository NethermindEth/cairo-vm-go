package utils

import "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"

type SortFelt []*fp.Element

func (s SortFelt) Len() int {
	return len(s)
}

func (s SortFelt) Less(i, j int) bool {
	return s[i].Cmp(s[j]) < 0
}

func (s SortFelt) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
