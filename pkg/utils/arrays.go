package utils

import (
	"fmt"
)

func Reverse[T any](a []T) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func Pop[T any](a *[]T) (T, error) {
	if len(*a) == 0 {
		var zeroValue T
		return zeroValue, fmt.Errorf("cannot pop from an empty slice")
	}

	v := (*a)[len(*a)-1]
	*a = (*a)[:len(*a)-1]
	return v, nil
}

func Contains[T comparable](a []T, v T) bool {
	for _, e := range a {
		if e == v {
			return true
		}
	}
	return false
}
