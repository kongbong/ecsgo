package sparseMap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMap(t *testing.T) {
	set := NewHash[int, int]()
	set.InsertVal(5, 100)
	set.InsertVal(1, 1000)

	assert.Equal(t, 100, *set.Find(5))
	assert.Equal(t, 1000, *set.Find(1))

	set.Erase(5)
	assert.Nil(t, set.Find(5))
	set.InsertVal(3, 200)
	assert.Equal(t, 200, *set.Find(3))
}

func TestHashMapRemoveIterate(t *testing.T) {
	set := NewHash[int, int]()
	set.InsertVal(5, 100)
	set.Erase(5)

	assert.Zero(t, len(set.Iterate()))

	for i := 0; i < 10; i++ {
		assert.True(t, set.InsertVal(i, i*100))
	}
	set.Erase(3)
	set.Erase(5)
	assert.Equal(t, 8, len(set.Iterate()))
}

func BenchmarkHashMap(b *testing.B) {
	set := NewHash[int, int]()
	for n := 0; n < b.N; n++ {
		set.InsertVal(n, n * 100)
	}
	for n := 0; n < b.N; n++ {
		set.Find(n)
	}
	for n := 0; n < b.N; n++ {
		set.Erase(n)
	}
}