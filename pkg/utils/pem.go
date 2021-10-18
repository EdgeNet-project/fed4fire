package utils

import (
	"encoding/pem"
)

// PEMDecodeMany decodes multiple PEM encoded blocks from a single byte slice.
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

// PEMEncodeMany PEM encodes multiple byte slices to a single byte slice.
func PEMEncodeMany(data [][]byte, pemType string) []byte {
	pemEncoded := make([]byte, 0)
	for _, d := range data {
		pemEncoded = append(pemEncoded, pem.EncodeToMemory(&pem.Block{Type: pemType, Bytes: d})...)
	}
	return pemEncoded
}
