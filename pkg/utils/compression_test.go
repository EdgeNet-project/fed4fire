package utils

import (
	"bytes"
	"testing"
)

func TestZlibBase64(t *testing.T) {
	want := []byte("Hello World")
	got := DecompressZlibBase64(CompressZlibBase64(want))
	if !bytes.Equal(got, want) {
		t.Errorf("DecompressZlibBase64(CompressZlibBase64()) = %s; want %s", got, want)
	}
}
