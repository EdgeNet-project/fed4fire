package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArrayFlags(t *testing.T) {
	var flag ArrayFlags
	assert.Nil(t, flag.Set("first"))
	assert.Nil(t, flag.Set("second"))
	assert.Equal(t, "first second", flag.String())
}
