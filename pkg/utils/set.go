package utils

type Set[T comparable] struct {
	set map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{set: make(map[T]struct{})}
}

func (s Set[T]) Add(v ...T) {
	for _, t := range v {
		s.set[t] = struct{}{}
	}
}

func (s Set[T]) Contains(v T) bool {
	_, ok := s.set[v]
	return ok
}

func (s Set[T]) All() []T {
	var tmp []T
	for t := range s.set {
		tmp = append(tmp, t)
	}
	return tmp
}

func (s Set[T]) Remove(v ...T) {
	for _, t := range v {
		delete(s.set, t)
	}
}
