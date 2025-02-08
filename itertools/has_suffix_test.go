package itertools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasSuffix(t *testing.T) {
	suffix := "bar"

	hasSuffixBar := HasSuffix(suffix)

	assert.True(t, hasSuffixBar("foobar"))
	assert.False(t, hasSuffixBar("foobarbaz"))
}
