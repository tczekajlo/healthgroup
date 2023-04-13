package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAtLevel(t *testing.T) {
	t.Parallel()

	logger, err := NewAtLevel("DEBUG")

	assert.NotNil(t, logger)
	assert.Empty(t, err)
}
