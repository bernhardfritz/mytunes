package itertools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNot(t *testing.T) {
	isOdd := func(a int) bool {
		return a%2 == 1
	}

	isEven := Not(isOdd)

	assert.False(t, isEven(1))
	assert.True(t, isEven(2))
	assert.False(t, isEven(3))
}
