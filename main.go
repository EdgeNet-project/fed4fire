package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"github.com/EdgeNet-project/fed4fire/pkg/service"
	"github.com/divan/gorilla-xmlrpc/xml"
	"github.com/gorilla/rpc"
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
var serverAddr string
var serverCertFile string
var serverKeyFile string
var trustedRootCerts arrayFlags

func main() {
	// TODO: usage
	flag.BoolVar(&showHelp, "help", false, "show this message")
	flag.StringVar(&serverAddr, "serverAddr", "localhost:9443", "")
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

	s := &service.Service{
		AbsoluteURL: serverAddr,
	}

	RPC := rpc.NewServer()
	xmlrpcCodec := xml.NewCodec()
	RPC.RegisterCodec(xmlrpcCodec, "text/xml")
	err := RPC.RegisterService(s, "")
	check(err)

	// https://github.com/divan/gorilla-xmlrpc/issues/14
	// TODO: Custom codec that append service name instead?
	xmlrpcCodec.RegisterAlias("GetVersion", "Service.GetVersion")

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      serverAddr,
		Handler:   RPC,
		TLSConfig: tlsConfig,
	}

	log.Printf("listening on %s", serverAddr)
	err = server.ListenAndServeTLS(serverCertFile, serverKeyFile)
	check(err)
}
