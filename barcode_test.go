package scale

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTerminator(t *testing.T) {
	assert.True(t, isTerminator([]byte{0, 0, 0, 0, 0, 0, 0, 0}))
	assert.False(t, isTerminator([]byte{0, 0, 0, 0, 1, 0, 0, 0}))
}

func TestIsShift(t *testing.T) {
	assert.True(t, isShift([]byte{2, 0, 0, 0, 0, 0, 0, 0}))
	assert.False(t, isShift([]byte{0, 0, 0, 0, 1, 0, 0, 0}))
}

func TestParseBuffer(t *testing.T) {
	r, err := ParseBuffer([]byte{})
	assert.Equal(t, r, "")
	assert.NotNil(t, err)

	r, err = ParseBuffer([]byte{1})
	assert.Equal(t, r, "")
	assert.NotNil(t, err)

	r, err = ParseBuffer([]byte{0, 0, 30, 0, 0, 0, 0, 0})
	assert.Equal(t, r, "1")
	assert.Nil(t, err)

	r, err = ParseBuffer([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	assert.Equal(t, r, TERMINATOR_STR)
	assert.Nil(t, err)

	r, err = ParseBuffer([]byte{2, 0, 0, 0, 0, 0, 0, 0})
	assert.Equal(t, r, SHIFT_KEY_STR)
	assert.Nil(t, err)

	r, err = ParseBuffer([]byte{0, 0, 30, 0, 0, 0, 0, 0})
	assert.Equal(t, r, "1")
	assert.Nil(t, err)
}
