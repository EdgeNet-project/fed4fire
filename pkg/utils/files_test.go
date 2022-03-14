package utils

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestWriteRemoveFiles(t *testing.T) {
	data := [][]byte{[]byte("Hello"), []byte("World")}
	names, err := WriteTempFiles(data)
	assert.Nil(t, err)
	assert.Len(t, names, len(data))
	for i, name := range names {
		b, _ := ioutil.ReadFile(name)
		assert.Equal(t, data[i], b)
	}
	RemoveFiles(names)
	for _, name := range names {
		assert.NoFileExists(t, name)
	}
}
