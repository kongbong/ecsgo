package ecsgo

import (
	"sync"
	"sync/atomic"
)

// Name is const string map for ecsgo component string variable
// Name map is always increasing not decreasing, so should be careful about memory usage.
type Name struct {
	idx int32
}

var idxMap sync.Map
var strMap sync.Map
var lastIdx int32

func NewName(str string) Name {
	if val, ok := strMap.Load(str); ok {
		return Name{idx: val.(int32)}
	}
	idx := atomic.AddInt32(&lastIdx, 1)
	idxMap.Store(idx, str)
	strMap.Store(str, idx)
	return Name{idx: idx}
}

func (n Name) String() string {
	if val, ok := idxMap.Load(n.idx); ok {
		return val.(string)
	} else {
		panic("no string of idx")
	}
}
