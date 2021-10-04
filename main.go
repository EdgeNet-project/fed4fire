package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	"github.com/EdgeNet-project/fed4fire/pkg/service"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"github.com/gorilla/rpc"
	"github.com/maxmouchet/gorilla-xmlrpc/xml"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"net/http"
	"os"
)

func logRequest(i *rpc.RequestInfo) {
	klog.InfoS(
		"Received XML-RPC request",
		"proto", i.Request.Proto,
		"method", i.Request.Method,
		"uri", i.Request.RequestURI,
		"rpc-method", i.Method,
		"user-agent", i.Request.UserAgent(),
		"request", utils.RequestId(i.Request),
	)
}

var showHelp bool
var authorityName string
var kubeconfigFile string
var serverAddr string
var serverCertFile string
var serverKeyFile string
var trustedRootCerts utils.ArrayFlags

func main() {
	klog.InitFlags(nil)
	flag.BoolVar(&showHelp, "help", false, "show this message")
	flag.StringVar(&authorityName, "authorityName", "edge-net.org", "authority name to use in URNs")
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
		klog.InfoS("Loading trusted certificate", "file", file)
		caCert, err := ioutil.ReadFile(file)
		utils.Check(err)
		caCertPool.AppendCertsFromPEM(caCert)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	utils.Check(err)

	edgeclient, err := versioned.NewForConfig(config)
	utils.Check(err)

	kubeclient, err := kubernetes.NewForConfig(config)
	utils.Check(err)

	// TODO: Read from YAML file
	containerImages := map[string]string{
		"ubuntu2004": "docker.io/library/ubuntu:20.04",
	}

	s := &service.Service{
		AbsoluteURL:     fmt.Sprintf("https://%s", serverAddr),
		AuthorityName:   authorityName,
		ContainerImages: containerImages,
		// TODO: From flag.
		ParentNamespace:  "lip6-lab-fed4fire-dev",
		EdgenetClient:    edgeclient,
		KubernetesClient: kubeclient,
	}

	xmlrpcCodec := xml.NewCodec()
	xmlrpcCodec.SetPrefix("Service.")

	RPC := rpc.NewServer()
	RPC.RegisterBeforeFunc(logRequest)
	RPC.RegisterCodec(xmlrpcCodec, "text/xml")
	utils.Check(RPC.RegisterService(s, ""))

	tlsConfig := &tls.Config{
		ClientCAs: caCertPool,
		//ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      serverAddr,
		Handler:   RPC,
		TLSConfig: tlsConfig,
	}

	klog.InfoS("Listening", "address", serverAddr)
	utils.Check(server.ListenAndServeTLS(serverCertFile, serverKeyFile))
}
