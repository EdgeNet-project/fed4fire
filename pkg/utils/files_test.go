package utils

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestWriteRemoveFiles(t *testing.T) {
	data := [][]byte{[]byte("Hello"), []byte("World")}
	names, err := WriteTempFiles(data)
	if err != nil {
		t.Errorf("WriteTempFiles() = %s; want nil", err)
	}
	if len(names) != len(data) {
		t.Errorf("len(names) = %d; want %d", len(names), len(data))
	}
	for i, name := range names {
		b, _ := ioutil.ReadFile(name)
		if !bytes.Equal(b, data[i]) {
			t.Errorf("ReadFile() = %s; want %s", b, data[i])
		}
	}
	RemoveFiles(names)
	for _, name := range names {
		if _, err := os.Stat(name); err == nil {
			t.Errorf("%s should not exist", name)
		}
	}
}
