package utils

import "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"

type SortFelt []fp.Element

func (s SortFelt) Len() int {
	return len(s)
}

func (s SortFelt) Less(i, j int) bool {
	feltOne := &s[i]
	feltTwo := &s[j]
	return feltOne.Cmp(feltTwo) < 0
}

func (s SortFelt) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
