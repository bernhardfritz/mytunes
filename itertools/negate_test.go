package itertools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNegate(t *testing.T) {
	isOdd := func(a int) bool {
		return a%2 == 1
	}

	isEven := Negate(isOdd)

	assert.False(t, isEven(1))
	assert.True(t, isEven(2))
	assert.False(t, isEven(3))
}
