package itertools

import (
	"slices"
	"testing"
)

func isOdd(a int) bool {
	return a%2 == 1
}

func TestFilter(t *testing.T) {
	want := []int{1, 3}
	if got := slices.Collect(Filter(isOdd, slices.Values([]int{1, 2, 3}))); !slices.Equal(got, want) {
		t.Errorf("Filter(isOdd, slices.Values([]int{1, 2, 3})) = %v, want %v", got, want)
	}
}
