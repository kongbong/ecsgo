package ecsgo

import (
	"reflect"
	"sort"
	"sync"
)

var typeMap map[reflect.Type]typeId = make(map[reflect.Type]typeId)
var lastTypeId int
var mtx sync.Mutex

type typeId struct {
	id int
	tp reflect.Type
}
type typeIdSlize []typeId

func (s typeIdSlize) Len() int {
	return len(s)
}
func (s typeIdSlize) Less(i, j int) bool {
	return s[i].id < s[j].id
}
func (s typeIdSlize) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortTypes(types []reflect.Type) {
	mtx.Lock()
	defer mtx.Unlock()
	var s typeIdSlize
	for _, t := range types {
		if typeid, ok := typeMap[t]; ok {
			s = append(s, typeid)
		} else {
			lastTypeId++
			typeMap[t] = typeId{lastTypeId, t}
			s = append(s, typeId{lastTypeId, t})
		}
	}
	sort.Sort(s)

	for i := 0; i < len(types); i++ {
		types[i] = s[i].tp
	}
}
