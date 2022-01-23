package sparseSet

type Set[T any] struct {
	dense []T
	sparse []uint32
	freelist []uint32
	autoincresing bool
}

func New[T any](maxValue uint32) *Set[T] {
	return &Set[T]{
		sparse: make([]uint32, maxValue),
		autoincresing: false,
	}
}

func NewAutoIncresing[T any](maxValue uint32) *Set[T] {
	return &Set[T]{
		sparse: make([]uint32, maxValue),
		autoincresing: true,
	}
}

func (s *Set[T]) NewSize(sz uint32) {
	if len(s.sparse) >= int(sz) {
		panic("only increasing is possible")
	}

	newSparse := make([]uint32, sz)
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
			newSize := uint32(len(s.sparse)*2)
			if newSize < id {
				newSize = id+1
			}
			s.NewSize(newSize)
		} else {
			return false
		}
	}
	if s.sparse[id] != 0 {
		// already inserted
		return false
	}
	if len(s.freelist) > 0 {
		// recycle
		last := s.freelist[len(s.freelist)-1]
		s.freelist = s.freelist[:len(s.freelist)-1]
		s.sparse[id] = last
		s.dense[last-1] = *val
		return true
	}
	s.dense = append(s.dense, *val)
	s.sparse[id] = uint32(len(s.dense))
	return true
}

func (s *Set[T]) Find(id uint32) *T {
	idx := s.sparse[id]
	if idx == 0 || int(idx) > len(s.dense) {
		// not inserted
		return nil
	}
	return &s.dense[idx-1]
}

func (s *Set[T]) Erase(id uint32) {
	idx := s.sparse[id]
	if idx == 0 || int(idx) > len(s.dense) {
		// already removed
		return
	}
	s.sparse[id] = 0
	s.freelist = append(s.freelist, idx)
}

func (s *Set[T]) Clear() {
	s.dense = s.dense[:0]
	s.freelist = s.freelist[:0]
}

func (s *Set[T]) Iterate() []T {
	return s.dense
}