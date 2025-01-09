package itertools

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	isOdd := func(a int) bool {
		return a%2 == 1
	}
	iterator := slices.Values([]int{1, 2, 3})

	filtered := slices.Collect(Filter(isOdd, iterator))

	assert.Equal(t, []int{1, 3}, filtered)
}
