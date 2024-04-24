package utils

func Reverse[T any](a []T) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func Pop[T any](a *[]T) T {
	v := (*a)[len(*a)-1]
	*a = (*a)[:len(*a)-1]
	return v
}
