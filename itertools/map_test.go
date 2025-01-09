package itertools

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	addOne := func(a int) int {
		return a + 1
	}
	iterator := slices.Values([]int{1, 2, 3})

	mapped := slices.Collect(Map(addOne, iterator))

	assert.Equal(t, mapped, []int{2, 3, 4})
}
