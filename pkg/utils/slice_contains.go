package utils

//func SliceContains[T any](s []T, eq func(T, T) bool, v T) bool {
//	for _, t := range s {
//		if eq(t, v) {
//			return true
//		}
//	}
//	return false
//}
//
//// Index return index of matched value; -1 if not matched
//func Index[T any](s []T, eq func(T, T) bool, v T) int {
//	for i, t := range s {
//		if eq(t, v) {
//			return i
//		}
//	}
//	return -1
//}
//
//func AddIfNotContains[T any](s []T, eq func(T, T) bool, v ...T) []T {
//	if len(s) == 0 {
//		return append(s, v...)
//	}
//	for _, t := range v {
//		if SliceContains[T](s, eq, t) {
//			continue
//		}
//		s = append(s, t)
//	}
//	return s
//}

type Slice[T any] struct {
	eq func(a, b T) bool
}

func NewSlice[T any](equalFunc func(a, b T) bool) Slice[T] {
	return Slice[T]{eq: equalFunc}
}

func (s Slice[T]) Contains(slice []T, v T) bool {
	for _, t := range slice {
		if s.eq(t, v) {
			return true
		}
	}
	return false
}

func (s Slice[T]) Index(slice []T, v T) int {
	for i, t := range slice {
		if s.eq(t, v) {
			return i
		}
	}
	return -1
}

func (s Slice[T]) AppendIfNotExist(slice []T, v ...T) []T {
	if len(slice) == 0 {
		return append(slice, v...)
	}
	for _, t := range v {
		if s.Contains(slice, t) {
			continue
		}
		slice = append(slice, t)
	}
	return slice
}
