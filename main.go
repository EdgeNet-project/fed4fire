package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// https://stackoverflow.com/a/28323276
type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, " ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var showHelp bool
var serverCertFile string
var serverKeyFile string
var trustedRootCerts arrayFlags

func main() {
	flag.BoolVar(&showHelp, "help", false, "show this message")
	flag.StringVar(&serverCertFile, "serverCert", "", "")
	flag.StringVar(&serverKeyFile, "serverKey", "", "")
	flag.Var(&trustedRootCerts, "trustedRootCert", "")
	flag.Parse()

	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	caCertPool := x509.NewCertPool()
	for _, file := range trustedRootCerts {
		log.Printf("loading trusted certificate %s", file)
		caCert, err := ioutil.ReadFile(file)
		check(err)
		caCertPool.AppendCertsFromPEM(caCert)
	}

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      ":9443",
		TLSConfig: tlsConfig,
	}

	err := server.ListenAndServeTLS(serverCertFile, serverKeyFile)
	check(err)
}
