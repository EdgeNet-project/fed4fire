package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeys(t *testing.T) {
	m := map[string]string{"Hello": "A", "World": "B"}
	assert.ElementsMatch(t, Keys(m), []string{"Hello", "World"})
}
