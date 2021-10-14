package utils

import (
	"encoding/pem"
	"os"
)

const (
	PEMBlockTypeCertificate = "CERTIFICATE"
	PEMBlockTypeRSA         = "RSA PRIVATE KEY"
)

// WriteTempFile writes data to a temporary file and returns its name.
// It is the caller's responsibility to remove the file when it is no longer needed.
func WriteTempFile(data []byte) (string, error) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	_, err = file.Write(data)
	if err != nil {
		RemoveFile(file.Name())
		return "", err
	}
	err = file.Close()
	if err != nil {
		RemoveFile(file.Name())
		return "", err
	}
	return file.Name(), nil
}

// WriteTempFiles writes data to temporary files and returns their names.
// It is the caller's responsibility to remove the files when they are no longer needed.
func WriteTempFiles(data [][]byte) ([]string, error) {
	names := make([]string, 0)
	for _, d := range data {
		file, err := WriteTempFile(d)
		if err != nil {
			RemoveFiles(names)
			return nil, err
		}
		names = append(names, file)
	}
	return names, nil
}

// WriteTempFilePem writes PEM encoded data to a temporary file and return its name.
// It is the caller's responsibility to remove the file when it is no longer needed.
func WriteTempFilePem(data []byte, pemType string) (string, error) {
	b := pem.EncodeToMemory(&pem.Block{Type: pemType, Bytes: data})
	return WriteTempFile(b)
}

// WriteTempFilePems writes multiple PEM encoded data to a temporary file and return its name.
// It is the caller's responsibility to remove the file when it is no longer needed.
func WriteTempFilePems(data [][]byte, pemType string) (string, error) {
	return WriteTempFile(PEMEncodeMany(data, pemType))
}

// WriteTempFilesPem write PEM encoded data to temporary files and returns their names.
// It is the caller's responsibility to remove the files when they are no longer needed.
func WriteTempFilesPem(data [][]byte, pemType string) ([]string, error) {
	bs := make([][]byte, 0)
	for _, d := range data {
		bs = append(bs, pem.EncodeToMemory(&pem.Block{Type: pemType, Bytes: d}))
	}
	return WriteTempFiles(bs)
}

// RemoveFile removes a file, ignoring all errors.
func RemoveFile(name string) {
	_ = os.Remove(name)
}

// RemoveFiles removes multiple files, ignoring all errors.
func RemoveFiles(names []string) {
	for _, name := range names {
		RemoveFile(name)
	}
}
