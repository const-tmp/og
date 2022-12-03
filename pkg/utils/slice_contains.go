package utils

func SliceContains[T any](s []T, eq func(T, T) bool, v T) bool {
	for _, t := range s {
		if eq(t, v) {
			return true
		}
	}
	return false
}

func AddIfNotContains[T any](s []T, eq func(T, T) bool, v ...T) []T {
	for _, t := range v {
		if SliceContains[T](s, eq, t) {
			continue
		}
		s = append(s, t)
	}
	return s
}
