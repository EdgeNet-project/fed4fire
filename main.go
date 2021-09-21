package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"github.com/EdgeNet-project/fed4fire/pkg/service"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"github.com/gorilla/handlers"
	"github.com/gorilla/rpc"
	"github.com/maxmouchet/gorilla-xmlrpc/xml"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
)

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
var trustedRootCerts utils.ArrayFlags

func main() {
	flag.BoolVar(&showHelp, "help", false, "show this message")
	flag.StringVar(&kubeconfigFile, "kubeconfig", "", "path to the kubeconfig file used to communicate with the Kubernetes API")
	flag.StringVar(&serverAddr, "serverAddr", "localhost:9443", "host:port on which to listen")
	flag.StringVar(&serverCertFile, "serverCert", "", "path to the server TLS certificate")
	flag.StringVar(&serverKeyFile, "serverKey", "", "path to the server TLS key")
	flag.Var(&trustedRootCerts, "trustedRootCert", "path to a trusted root certificate for authenticating user; can be specified multiple times")
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
		AbsoluteURL:      serverAddr,
		URN:              "urn:publicid:IDN+edge-net.org+authority+am",
		KubernetesClient: clientset,
	}

	xmlrpcCodec := xml.NewCodec()
	xmlrpcCodec.SetPrefix("Service.")

	RPC := rpc.NewServer()
	RPC.RegisterBeforeFunc(logRequest)
	RPC.RegisterCodec(xmlrpcCodec, "text/xml")
	check(RPC.RegisterService(s, ""))

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
