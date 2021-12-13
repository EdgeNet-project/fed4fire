package utils

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"io/ioutil"
)

func CompressZlibBase64(data []byte) string {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		panic(err)
	}
	err = w.Close()
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func DecompressZlibBase64(s string) []byte {
	c, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	r, err := zlib.NewReader(bytes.NewReader(c))
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	err = r.Close()
	if err != nil {
		panic(err)
	}
	return b
}
