package ecsgo

import "testing"

func TestRegistry_Tick_sysDeltaSecondsRace(t *testing.T) {
	registry := New()
	AddSystem(registry, OnTick, func(r *Registry) {})
	AddSystem(registry, OnTick, func(r *Registry) {})
	registry.Tick(1.0)
}
