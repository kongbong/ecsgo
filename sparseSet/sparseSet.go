package sparseSet

import "log"

// Set sparseSet
type Set[T any] struct {
	denseMap      []uint32
	dense         []T
	sparse        []uint32
	autoincresing bool
}

func New[T any](maxValue uint32) *Set[T] {
	return &Set[T]{
		sparse: make([]uint32, maxValue+1),
		autoincresing: false,
	}
}

func NewAutoIncresing[T any](maxValue uint32) *Set[T] {
	return &Set[T]{
		sparse: make([]uint32, maxValue+1),
		autoincresing: true,
	}
}

func (s *Set[T]) newMaxValue(maxValue uint32) {
	if len(s.sparse) >= int(maxValue+1) {
		panic("only increasing is possible")
	}

	newSparse := make([]uint32, maxValue+1)
	copy(newSparse, s.sparse[:])
	s.sparse = newSparse
}

// for primitive type values
func (s *Set[T]) InsertVal(id uint32, val T) bool {
	return s.Insert(id, &val)
}

func (s *Set[T]) Insert(id uint32, val *T) bool {
	if int(id) >= len(s.sparse) {
		if s.autoincresing {
			newMaxValue := uint32(len(s.sparse)*2)
			if newMaxValue < id {
				newMaxValue = id+1
			}
			s.newMaxValue(newMaxValue)
		} else {
			log.Println("exceeing maxvalue")
			return false
		}
	}
	if s.sparse[id] != 0 {
		// already inserted
		log.Println("already inserted", id)
		return false
	}

	s.dense = append(s.dense, *val)
	s.denseMap = append(s.denseMap, id)
	s.sparse[id] = uint32(len(s.dense))
	return true
}

func (s *Set[T]) Find(id uint32) *T {
	if int(id) >= len(s.sparse) {
		// exceed maxValue
		return nil
	}
	idx := s.sparse[id]
	if idx == 0 || int(idx) > len(s.dense) {
		// not inserted
		return nil
	}
	return &s.dense[idx-1]
}

func (s *Set[T]) Erase(id uint32) {
	if s.Find(id) == nil {
		// not inserted
		return
	}

	idx := s.sparse[id]
	s.sparse[id] = 0
	last := s.dense[len(s.dense)-1]
	lastSparse := s.denseMap[len(s.denseMap)-1]

	if int(idx) < len(s.dense) {
		// removed last element, don't need to swap
		s.dense[idx-1] = last
		s.denseMap[idx-1] = lastSparse
		s.sparse[lastSparse] = idx
	}
	
	s.dense = s.dense[:len(s.dense)-1]
	s.denseMap = s.denseMap[:len(s.denseMap)-1]
}

func (s *Set[T]) Clear() {
	s.dense = s.dense[:0]
	s.denseMap = s.denseMap[:0]
}

func (s *Set[T]) Iterate() []T {
	return s.dense
}