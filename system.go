package ecsgo

import (
	"reflect"
	"slices"
	"time"
)

type ExecutionContext struct {
	registry  *Registry
	deltaTime time.Duration

	queryResults []*QueryResult
}

type QueryResult struct {
	query         *Query
	archeTypeList []*ArcheType
}

// Component Query
type Query struct {
	includeComponents    []reflect.Type
	excludeComponents    []reflect.Type
	optionalComponents   []reflect.Type
	readonlyComponents   []reflect.Type
	atleastOneComponents [][]reflect.Type

	interestedArcheTypeList []*ArcheType
}

type SystemFn func(ctx *ExecutionContext) error

type System struct {
	registry *Registry
	name     string
	priority int
	fn       SystemFn

	// query
	queries []*Query
}

func newSystem(registry *Registry, name string, priority int, fn SystemFn) *System {
	return &System{
		registry: registry,
		name:     name,
		priority: priority,
		fn:       fn,
	}
}

func (s *System) execute(deltaTime time.Duration) error {
	ctx := &ExecutionContext{
		registry:  s.registry,
		deltaTime: deltaTime,
	}
	for _, q := range s.queries {
		qr := &QueryResult{
			query:         q,
			archeTypeList: q.interestedArcheTypeList,
		}
		ctx.queryResults = append(ctx.queryResults, qr)
	}
	return s.fn(ctx)
}

func (s *System) GetName() string {
	return s.name
}

func (s *System) GetPriority() int {
	return s.priority
}

func (s *System) NewQuery() *Query {
	q := &Query{}
	s.queries = append(s.queries, q)
	return q
}

func AddReadWriteComponent[T any](q *Query) {
	var t T
	ty := reflect.TypeOf(t)
	if q.hasComponent(ty) {
		// already added
		return
	}
	q.includeComponents = append(q.includeComponents, ty)
}

func AddReadonlyComponent[T any](q *Query) {
	var t T
	ty := reflect.TypeOf(t)
	if q.hasComponent(ty) {
		// already added
		return
	}
	q.includeComponents = append(q.includeComponents, ty)
	q.readonlyComponents = append(q.readonlyComponents, ty)
}

func AddExcludeComponent[T any](q *Query) {
	var t T
	ty := reflect.TypeOf(t)
	if q.hasComponent(ty) {
		// already added
		return
	}
	q.excludeComponents = append(q.excludeComponents, ty)
}

func AddOptionalReadWriteComponent[T any](q *Query) {
	var t T
	ty := reflect.TypeOf(t)
	if q.hasComponent(ty) {
		// already added
		return
	}
	q.optionalComponents = append(q.optionalComponents, ty)
}

func AddOptionalReadonlyComponent[T any](q *Query) {
	var t T
	ty := reflect.TypeOf(t)
	if q.hasComponent(ty) {
		// already added
		return
	}
	q.optionalComponents = append(q.optionalComponents, ty)
	q.readonlyComponents = append(q.readonlyComponents, ty)
}

func (q *Query) AtLeastOneOfThem(tps []reflect.Type) {
	q.atleastOneComponents = append(q.atleastOneComponents, tps)
}

func (q *Query) AtLeastOneOfThemReadonly(tps []reflect.Type) {
	q.atleastOneComponents = append(q.atleastOneComponents, tps)
	q.readonlyComponents = append(q.readonlyComponents, tps...)
}

func (s *System) hasComponent(ty reflect.Type) bool {
	for _, q := range s.queries {
		if q.hasComponent(ty) {
			return true
		}
	}
	return false
}

func (q *Query) hasComponent(ty reflect.Type) bool {
	if q.isInterestComponent(ty) {
		return true
	}
	if slices.Contains(q.excludeComponents, ty) {
		return true
	}
	return false
}

func (s *System) isInterestComponent(ty reflect.Type) bool {
	for _, q := range s.queries {
		if q.isInterestComponent(ty) {
			return true
		}
	}
	return false
}

func (q *Query) isInterestComponent(ty reflect.Type) bool {
	if slices.Contains(q.includeComponents, ty) {
		return true
	}
	if slices.Contains(q.optionalComponents, ty) {
		return true
	}
	for _, atleast := range q.atleastOneComponents {
		if slices.Contains(atleast, ty) {
			return true
		}
	}
	return false
}

func (s *System) getInterestComponentCount() int {
	cnt := 0
	for _, q := range s.queries {
		cnt += q.getInterestComponentCount()
	}
	return cnt
}

func (q *Query) getInterestComponentCount() int {
	cnt := len(q.includeComponents) + len(q.optionalComponents)
	for _, atleast := range q.atleastOneComponents {
		cnt += len(atleast)
	}
	return cnt
}

func (s *System) dependent(other *System) bool {
	for _, q := range s.queries {
		for _, otherQ := range other.queries {
			if q.dependent(otherQ) {
				return true
			}
		}
	}
	return false
}

func (q *Query) dependent(other *Query) bool {
	for _, t := range q.includeComponents {
		if other.isInterestComponent(t) {
			if !slices.Contains(q.readonlyComponents, t) || !slices.Contains(other.readonlyComponents, t) {
				return true
			}
		}
	}
	for _, t := range q.optionalComponents {
		if other.isInterestComponent(t) {
			if !slices.Contains(q.readonlyComponents, t) || !slices.Contains(other.readonlyComponents, t) {
				return true
			}
		}
	}
	for _, atleast := range q.atleastOneComponents {
		for _, t := range atleast {
			if other.isInterestComponent(t) {
				if !slices.Contains(q.readonlyComponents, t) || !slices.Contains(other.readonlyComponents, t) {
					return true
				}
			}
		}
	}
	return false
}

func (s *System) addArcheTypeIfInterest(archeType *ArcheType) bool {
	var added bool
	for _, q := range s.queries {
		if q.addArcheTypeIfInterest(archeType) {
			added = true
		}
	}
	return added
}

func (q *Query) addArcheTypeIfInterest(archeType *ArcheType) bool {
	for _, t := range q.includeComponents {
		if !archeType.hasComponent(t) {
			return false
		}
	}
	for _, t := range q.excludeComponents {
		if archeType.hasComponent(t) {
			return false
		}
	}
	for _, atleast := range q.atleastOneComponents {
		found := false
		for _, t := range atleast {
			if archeType.hasComponent(t) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	q.interestedArcheTypeList = append(q.interestedArcheTypeList, archeType)
	return true
}

func (c *ExecutionContext) GetResgiry() *Registry {
	return c.registry
}

func (c *ExecutionContext) CreateEntity() EntityId {
	return c.registry.CreateEntity()
}

func (c *ExecutionContext) GetDeltaTime() time.Duration {
	return c.deltaTime
}

func (c *ExecutionContext) GetQueryResultCount() int {
	return len(c.queryResults)
}

func (c *ExecutionContext) GetQueryResult(idx int) *QueryResult {
	return c.queryResults[idx]
}

func (qr *QueryResult) GetArcheTypeCount() int {
	return len(qr.archeTypeList)
}

func (qr *QueryResult) GetArcheType(idx int) *ArcheType {
	return qr.archeTypeList[idx]
}

func (qr *QueryResult) ForeachEntities(fn func(accessor *ArcheTypeAccessor) error) error {
	for _, archeType := range qr.archeTypeList {
		if archeType.getEntityCount() == 0 {
			continue
		}
		err := archeType.Foreach(func(accessor *ArcheTypeAccessor) error {
			err := fn(accessor)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetComponent[T any](c *ExecutionContext, entityId EntityId) *T {
	var t T
	for i := 0; i < c.GetQueryResultCount(); i++ {
		qr := c.GetQueryResult(i)
		for j := 0; j < qr.GetArcheTypeCount(); j++ {
			a := qr.GetArcheType(j)
			if a == nil {
				continue
			}
			if a.hasComponent(reflect.TypeOf(t)) {
				valT := getArcheTypeComponent[T](a, entityId)
				if valT != nil {
					return valT
				}
			}
		}
	}
	return nil
}

func HasComponent[T any](c *ExecutionContext, entityId EntityId) bool {
	var t T
	for i := 0; i < c.GetQueryResultCount(); i++ {
		qr := c.GetQueryResult(i)
		for j := 0; j < qr.GetArcheTypeCount(); j++ {
			a := qr.GetArcheType(j)
			if a == nil {
				continue
			}
			if a.hasComponent(reflect.TypeOf(t)) {
				if a.hasEntity(entityId) {
					return true
				}
			}
		}
	}
	return false
}
