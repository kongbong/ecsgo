package sparseMap

import (
	"golang.org/x/exp/constraints"
	"log"
)

// Map sparseMap
type Map[K constraints.Integer, T any] struct {
	denseMap      []K
	dense         []T
	sparse        []K
	autoincresing bool
}

func New[K constraints.Integer, T any](maxValue K) *Map[K, T] {
	return &Map[K, T]{
		sparse: make([]K, maxValue+1),
		autoincresing: false,
	}
}

func NewAutoIncresing[K constraints.Integer, T any](maxValue K) *Map[K, T] {
	return &Map[K, T]{
		sparse: make([]K, maxValue+1),
		autoincresing: true,
	}
}

func (s *Map[K, T]) newMaxValue(maxValue K) {
	if len(s.sparse) >= int(maxValue+1) {
		panic("only increasing is possible")
	}

	newSparse := make([]K, maxValue+1)
	copy(newSparse, s.sparse[:])
	s.sparse = newSparse
}

// for primitive type values
func (s *Map[K, T]) InsertVal(id K, val T) bool {
	return s.Insert(id, &val)
}

func (s *Map[K, T]) Insert(id K, val *T) bool {
	if int(id) >= len(s.sparse) {
		if s.autoincresing {
			newMaxValue := K(len(s.sparse)*2)
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
	s.sparse[id] = K(len(s.dense))
	return true
}

func (s *Map[K, T]) Find(id K) *T {
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

func (s *Map[K, T]) Erase(id K) {
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

func (s *Map[K, T]) Clear() {
	s.dense = s.dense[:0]
	s.denseMap = s.denseMap[:0]
	s.sparse = make([]K, len(s.sparse))
}

func (s *Map[K, T]) Iterate() []T {
	return s.dense
}