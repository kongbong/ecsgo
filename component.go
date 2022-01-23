package ecsgo

// SetEntityComponent1 store entity data into storage
func SetEntityComponent1[T any](r *Registry, entity Entity, v *T) {
	t := getOrAddTable1[T](r)
	t.insert(entity, v)
}

// SetEntityComponent2 store entity data into storage
func SetEntityComponent2[T any, U any](r *Registry, entity Entity, v1 *T, v2 *U) {
	t := getOrAddTable2[T, U](r)
	t.insert(entity, v1, v2)
}