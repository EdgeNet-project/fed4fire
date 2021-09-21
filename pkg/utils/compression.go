package utils

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
)

func ZlibBase64(data []byte) string {
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
