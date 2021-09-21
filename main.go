package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"github.com/EdgeNet-project/fed4fire/pkg/service"
	"github.com/gorilla/handlers"
	"github.com/gorilla/rpc"
	"github.com/maxmouchet/gorilla-xmlrpc/xml"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

func logRequest(i *rpc.RequestInfo) {
	log.Printf("%s", i.Method)
}

var showHelp bool
var kubeconfigFile string
var serverAddr string
var serverCertFile string
var serverKeyFile string
var trustedRootCerts arrayFlags

func main() {
	// TODO: usage
	flag.BoolVar(&showHelp, "help", false, "show this message")
	flag.StringVar(&kubeconfigFile, "kubeconfig", "", "")
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

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	check(err)

	clientset, err := kubernetes.NewForConfig(config)
	check(err)

	s := &service.Service{
		AbsoluteURL:    serverAddr,
		URN:            "urn:publicid:IDN+edge-net.org+authority+am",
		KubernetesClient: clientset,
	}

	RPC := rpc.NewServer()
	xmlrpcCodec := xml.NewCodec()
	xmlrpcCodec.SetPrefix("Service.")
	RPC.RegisterCodec(xmlrpcCodec, "text/xml")
	RPC.RegisterBeforeFunc(logRequest)
	err = RPC.RegisterService(s, "")
	check(err)

	// TODO: Handle compression.

	tlsConfig := &tls.Config{
		ClientCAs: caCertPool,
		//ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      serverAddr,
		Handler:   handlers.LoggingHandler(os.Stdout, RPC),
		TLSConfig: tlsConfig,
	}

	log.Printf("listening on %s", serverAddr)
	err = server.ListenAndServeTLS(serverCertFile, serverKeyFile)
	check(err)
}
