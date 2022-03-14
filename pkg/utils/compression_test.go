package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZlibBase64(t *testing.T) {
	want := []byte("Hello World")
	got := DecompressZlibBase64(CompressZlibBase64(want))
	assert.Equal(t, want, got)
}
