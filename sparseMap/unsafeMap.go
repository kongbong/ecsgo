package sparseMap

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
	"constraints"
	"log"
)

// UnsafeMap sparseUnsafeMap
type UnsafeMap[K constraints.Integer] struct {
	valueSize     int
	denseMap      []K
	dense         unsafe.Pointer
	denseCap      int
	denseLen      int
	sparse        []K
	autoincresing bool
}

func NewUnsafe[K constraints.Integer](valueSize int, maxValue K) *UnsafeMap[K] {
	return &UnsafeMap[K]{
		valueSize: valueSize,
		sparse: make([]K, maxValue+1),
		autoincresing: false,
	}
}

func NewUnsafeAutoIncresing[K constraints.Integer](valueSize int, maxValue K) *UnsafeMap[K] {
	return &UnsafeMap[K]{
		valueSize: valueSize,
		sparse: make([]K, maxValue+1),
		autoincresing: true,
	}
}

// valueSize value size
func (s *UnsafeMap[K]) ValueSize() int {
	return s.valueSize
}

func (s *UnsafeMap[K]) newMaxValue(maxValue K) {
	if len(s.sparse) >= int(maxValue+1) {
		panic("only increasing is possible")
	}

	newSparse := make([]K, maxValue+1)
	copy(newSparse, s.sparse[:])
	s.sparse = newSparse
}

func get(arr unsafe.Pointer, idx int, valueSize int) unsafe.Pointer {
	return unsafe.Add(arr, valueSize * idx)
}

func set(arr unsafe.Pointer, idx int, val unsafe.Pointer, valueSize int) {
	C.memcpy(get(arr, idx, valueSize), val, C.size_t(valueSize))
}

func (s *UnsafeMap[K]) Insert(id K, val unsafe.Pointer) bool {
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

	if s.denseLen == s.denseCap {
		// need to increase dense array
		newCap := s.denseCap * 2 + 1
		newDense := unsafe.Pointer(C.malloc(C.size_t(newCap * s.valueSize)))
		C.memcpy(newDense, s.dense, C.size_t(s.denseCap * s.valueSize))
		if uintptr(s.dense) != uintptr(0) {
			C.free(s.dense)
		}		
		s.dense = newDense
		s.denseCap = newCap
	}
	
	set(s.dense, s.denseLen, val, s.valueSize)
	s.denseLen++
	s.denseMap = append(s.denseMap, id)
	s.sparse[id] = K(s.denseLen)
	return true
}

func (s *UnsafeMap[K]) Find(id K) (unsafe.Pointer, int) {
	if int(id) >= len(s.sparse) {
		// exceed maxValue
		return nil, 0
	}
	idx := s.sparse[id]
	if idx == 0 || int(idx) > s.denseLen {
		// not inserted
		return nil, 0
	}
	return get(s.dense, int(idx-1), s.valueSize), s.valueSize
}

func (s *UnsafeMap[K]) Erase(id K) {
	if int(id) >= len(s.sparse) {
		// exceed maxValue
		return
	}
	idx := s.sparse[id]
	if idx == 0 || int(idx) > s.denseLen {
		// not inserted
		return
	}

	s.sparse[id] = 0
	last := get(s.dense, s.denseLen-1, s.valueSize)
	lastSparse := s.denseMap[len(s.denseMap)-1]

	if int(idx) < s.denseLen {
		// removed last element, don't need to swap
		set(s.dense, int(idx-1), last, s.valueSize)
		s.denseMap[idx-1] = lastSparse
		s.sparse[lastSparse] = idx
	}
	
	s.denseLen--
	s.denseMap = s.denseMap[:len(s.denseMap)-1]
}

func (s *UnsafeMap[K]) Free() {
	C.free(s.dense)
	s.denseLen = 0
	s.denseCap = 0
	s.denseMap = s.denseMap[:0]
	s.sparse = make([]K, len(s.sparse))
}

type UnsafeIter struct {
	arr unsafe.Pointer
	valueSize int
	ln int
	idx int
}

func (i *UnsafeIter) IsNil() bool {
	return i.idx >= i.ln
}

func (i *UnsafeIter) Next() {
	i.idx++
}

func (i *UnsafeIter) Len() int {
	return i.ln
}

func (i *UnsafeIter) Get() unsafe.Pointer {
	return get(i.arr, i.idx, i.valueSize)
}

func (s *UnsafeMap[K]) Iterate() *UnsafeIter {
	return &UnsafeIter{ 
		arr: s.dense,
		valueSize: s.valueSize,
		ln: s.denseLen,
		idx: 0,
	}
}