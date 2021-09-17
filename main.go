package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	caCertPool := x509.NewCertPool()

	files, err := ioutil.ReadDir("trusted_roots")
	check(err)

	for _, file := range files {
		if !file.IsDir() {
			caCert, _ := ioutil.ReadFile(file.Name())
			caCertPool.AppendCertsFromPEM(caCert)
		}
	}

	// Append self-signed user authority root during dev.
	caCert, _ := ioutil.ReadFile("self_signed/ca-client.pem")
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs: caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      ":9443",
		TLSConfig: tlsConfig,
	}

	err = server.ListenAndServeTLS("self_signed/server.pem", "self_signed/server.key")
	check(err)
}