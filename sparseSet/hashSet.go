package sparseSet

import "log"

// Set sparseSet
type HashSet[K comparable, T any] struct {
	denseMap      []K
	dense         []T
	sparse        map[K]int
	autoincresing bool
}

func NewHash[K comparable, T any]() *HashSet[K, T] {
	return &HashSet[K, T]{
		sparse: make(map[K]int),
	}
}

// for primitive type values
func (s *HashSet[K, T]) InsertVal(id K, val T) bool {
	return s.Insert(id, &val)
}

func (s *HashSet[K, T]) Insert(id K, val *T) bool {
	
	if _, ok := s.sparse[id]; ok {
		// already inserted
		log.Println("already inserted", id)
		return false
	}

	s.dense = append(s.dense, *val)
	s.denseMap = append(s.denseMap, id)
	s.sparse[id] = len(s.dense)-1
	return true
}

func (s *HashSet[K, T]) Find(id K) *T {
	idx, ok := s.sparse[id]
	if !ok {
		// not inserted
		return nil
	}
	return &s.dense[idx]
}

func (s *HashSet[K, T]) Erase(id K) {
	idx, ok := s.sparse[id]
	if !ok {
		// not inserted
		return
	}
	delete(s.sparse, id)

	last := s.dense[len(s.dense)-1]
	lastSparse := s.denseMap[len(s.denseMap)-1]

	if idx < len(s.dense)-1 {
		// removed last element, don't need to swap
		s.dense[idx] = last
		s.denseMap[idx] = lastSparse
		s.sparse[lastSparse] = idx
	}
	
	s.dense = s.dense[:len(s.dense)-1]
	s.denseMap = s.denseMap[:len(s.denseMap)-1]
}

func (s *HashSet[K, T]) Clear() {
	s.dense = s.dense[:0]
	s.denseMap = s.denseMap[:0]
	s.sparse = make(map[K]int)
}

func (s *HashSet[K, T]) Iterate() []T {
	return s.dense
}