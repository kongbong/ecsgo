package sparseSet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	sparse := New[int](10)
	sparse.InsertVal(5, 100)
	sparse.InsertVal(1, 1000)

	assert.Equal(t, 100, *sparse.Find(5))
	assert.Equal(t, 1000, *sparse.Find(1))

	sparse.Erase(5)
	assert.Nil(t, sparse.Find(5))
	sparse.InsertVal(3, 200)
	assert.Equal(t, 200, *sparse.Find(3))

	sparse.newMaxValue(20)
	assert.Equal(t, 1000, *sparse.Find(1))
	assert.Equal(t, 200, *sparse.Find(3))
}

func TestAutoIncresingSet(t *testing.T) {
	sparse := NewAutoIncresing[int](5)

	for i := 0; i < 10; i++ {
		assert.True(t, sparse.InsertVal(uint32(i), i*100))
	}

	for i := 0; i < 10; i++ {
		assert.Equal(t, i*100, *sparse.Find(uint32(i)))
	}
}

func TestRemoveIterate(t *testing.T) {
	sparse := New[int](10)
	sparse.InsertVal(5, 100)
	sparse.Erase(5)

	assert.Zero(t, len(sparse.Iterate()))

	for i := 0; i < 10; i++ {
		assert.True(t, sparse.InsertVal(uint32(i), i*100))
	}
	sparse.Erase(3)
	sparse.Erase(5)
	assert.Equal(t, 8, len(sparse.Iterate()))
}
