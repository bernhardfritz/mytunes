package itertools

import (
	"slices"
	"testing"
)

func addOne(a int) int {
	return a + 1
}

func TestMap(t *testing.T) {
	want := []int{2, 3, 4}
	if got := slices.Collect(Map(addOne, slices.Values([]int{1, 2, 3}))); !slices.Equal(got, want) {
		t.Errorf("Map(addOne, slices.Values([]int{1, 2, 3})) = %v, want %v", got, want)
	}
}
