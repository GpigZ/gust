package iter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAny(t *testing.T) {
	var iter = FromVec([]int{1, 2, 3})
	if !iter.Any(func(x int) bool {
		return x > 1
	}) {
		t.Error("Any failed")
	}
}

func TestNextChunk(t *testing.T) {
	var iter = FromVec([]int{1, 2, 3})

	var chunk, ok = iter.NextChunk(2)
	assert.Equal(t, []int{1, 2}, chunk)
	assert.True(t, ok)

	chunk, ok = iter.NextChunk(2)
	assert.Equal(t, []int{3}, chunk)
	assert.False(t, ok)

	chunk, ok = iter.NextChunk(2)
	assert.Equal(t, []int{}, chunk)
	assert.False(t, ok)
}
