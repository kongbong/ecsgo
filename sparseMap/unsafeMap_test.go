package sparseMap

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestUnsafeMap(t *testing.T) {
	var val int = 100
	m := NewUnsafe[uint32](int(unsafe.Sizeof(val)), 10)
	defer m.Free()

	m.Insert(5, unsafe.Pointer(&val))
	val = 1000
	m.Insert(1, unsafe.Pointer(&val))

	p, _ := m.Find(5)
	assert.Equal(t, 100, *(*int)(p))
	p, _ = m.Find(1)
	assert.Equal(t, 1000, *(*int)(p))

	m.Erase(5)
	p, _ = m.Find(5)
	assert.Equal(t, uintptr(0), uintptr(p))
	val = 200
	m.Insert(3, unsafe.Pointer(&val))
	p, _ = m.Find(3)
	assert.Equal(t, 200, *(*int)(p))

	m.newMaxValue(20)
	p, _ = m.Find(1)
	assert.Equal(t, 1000, *(*int)(p))
	p, _ = m.Find(3)
	assert.Equal(t, 200, *(*int)(p))
}

func TestUnsafeAutoIncresingMap(t *testing.T) {
	var val int
	m := NewUnsafeAutoIncresing[uint32](int(unsafe.Sizeof(val)), 5)
	defer m.Free()

	for i := 0; i < 10; i++ {
		val = i * 100
		assert.True(t, m.Insert(uint32(i), unsafe.Pointer(&val)))
	}

	for i := 0; i < 10; i++ {
		p, _ := m.Find(uint32(i))
		assert.Equal(t, i*100, *(*int)(p))
	}
}

func TestUnsafeRemoveIterate(t *testing.T) {
	var val int = 100
	m := NewUnsafe[uint32](int(unsafe.Sizeof(val)), 10)
	defer m.Free()
	m.Insert(5, unsafe.Pointer(&val))
	m.Erase(5)

	assert.Zero(t, m.Iterate().Len())

	for i := 0; i < 10; i++ {
		val = i * 100
		assert.True(t, m.Insert(uint32(i), unsafe.Pointer(&val)))
	}
	m.Erase(3)
	m.Erase(5)
	assert.Equal(t, 8, m.Iterate().Len())
}

func BenchmarkUnsafeMap(b *testing.B) {
	var val int = 100
	m := NewUnsafe[int](int(unsafe.Sizeof(val)), b.N)
	defer m.Free()
	for n := 0; n < b.N; n++ {
		val = n * 100
		m.Insert(n, unsafe.Pointer(&val))
	}
	for n := 0; n < b.N; n++ {
		m.Find(n)
	}
	for n := 0; n < b.N; n++ {
		m.Erase(n)
	}
}
