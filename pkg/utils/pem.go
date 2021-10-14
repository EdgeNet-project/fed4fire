package utils

import (
	"encoding/pem"
)

func PEMDecodeMany(pemEncoded []byte) [][]byte {
	data := make([][]byte, 0)
	rest := pemEncoded
	var block *pem.Block
	for rest != nil {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		data = append(data, block.Bytes)
	}
	return data
}

func PEMEncodeMany(data [][]byte, pemType string) []byte {
	pemEncoded := make([]byte, 0)
	for _, d := range data {
		pemEncoded = append(pemEncoded, pem.EncodeToMemory(&pem.Block{Type: pemType, Bytes: d})...)
	}
	return pemEncoded
}
