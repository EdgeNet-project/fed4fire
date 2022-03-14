// This package implements the Fed4Fire Aggregate Manager API for EdgeNet/Kubernetes.
//
// Specifically, it implements the GENI AM API v3 as specified in
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3.
package main

import (
	"flag"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/gc"
	versioned "github.com/EdgeNet-project/fed4fire/pkg/generated/clientset/versioned"
	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
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
	"strings"
	"time"
)

var showHelp bool
var authorityName string
var containerImages utils.ArrayFlags
var containerCpuLimit string
var containerMemoryLimit string
var kubeconfigFile string
var namespace string
var serverAddr string
var trustedCerts utils.ArrayFlags

func beforeFunc(i *rpc.RequestInfo) {
	escapedCert := i.Request.Header.Get(constants.HttpHeaderCertificate)
	urn, err := utils.GetUserUrnFromEscapedCert(escapedCert)
	if err == nil {
		i.Request.Header.Set(constants.HttpHeaderUser, urn)
	} else {
		klog.ErrorS(err, "Failed to get user URN from header")
	}
	klog.InfoS(
		"Received XML-RPC request",
		"user-urn", i.Request.Header.Get(constants.HttpHeaderUser),
		"rpc-method", i.Method,
	)
}

func main() {
	klog.InitFlags(nil)
	flag.BoolVar(&showHelp, "help", false, "show this message")
	flag.StringVar(&authorityName, "authorityName", "example.org", "authority name to use in URNs")
	flag.Var(&containerImages, "containerImage", "name:image of a container image that can be deployed; can be specified multiple times")
	flag.StringVar(&containerCpuLimit, "containerCpuLimit", "2", "maximum amount of CPU that can be used by a container")
	flag.StringVar(&containerMemoryLimit, "containerMemoryLimit", "2Gi", "maximum amount of memory that can be used by a container")
	flag.StringVar(&kubeconfigFile, "kubeconfig", "", "path to the kubeconfig file used to communicate with the Kubernetes API")
	flag.StringVar(&namespace, "namespace", "", "kubernetes namespaces in which to create resources")
	flag.StringVar(&serverAddr, "serverAddr", "localhost:9443", "host:port on which to listen")
	flag.Var(&trustedCerts, "trustedCert", "path to a trusted certificate for authenticating users; can be specified multiple times")
	flag.Parse()

	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	utils.Check(err)

	f4fclient, err := versioned.NewForConfig(config)
	utils.Check(err)

	kubeclient, err := kubernetes.NewForConfig(config)
	utils.Check(err)

	authorityIdentifier := identifiers.Identifier{
		Authorities:  []string{authorityName},
		ResourceType: identifiers.ResourceTypeAuthority,
		ResourceName: "am",
	}

	containerImages_ := make(map[string]string)
	for _, s := range containerImages {
		arr := strings.SplitN(s, ":", 2)
		containerImages_[arr[0]] = arr[1]
		klog.InfoS("Parsed container image name", "name", arr[0], "image", arr[1])
	}

	trustedCerts_ := make([][]byte, 0)
	for _, s := range trustedCerts {
		b, err := ioutil.ReadFile(s)
		utils.Check(err)
		trustedCerts_ = append(trustedCerts_, utils.PEMDecodeMany(b)...)
	}

	s := &service.Service{
		// TODO: This is invalid with the reverse proxy, add absoluteUrl param?
		// and rename serverAddr to listenAddr?
		AbsoluteURL:          fmt.Sprintf("https://%s", serverAddr),
		AuthorityIdentifier:  authorityIdentifier,
		ContainerImages:      containerImages_,
		ContainerCpuLimit:    containerCpuLimit,
		ContainerMemoryLimit: containerMemoryLimit,
		Namespace:            namespace,
		TrustedCertificates:  trustedCerts_,
		Fed4FireClient:       f4fclient,
		KubernetesClient:     kubeclient,
	}

	xmlrpcCodec := xml.NewCodec()
	xmlrpcCodec.SetPrefix("Service.")

	RPC := rpc.NewServer()
	RPC.RegisterBeforeFunc(beforeFunc)
	RPC.RegisterCodec(xmlrpcCodec, "text/xml")
	utils.Check(RPC.RegisterService(s, ""))

	gc.GC{
		Fed4FireClient:   f4fclient,
		KubernetesClient: kubeclient,
		Interval:         5 * time.Second,
		Timeout:          30 * time.Second,
		Namespace:        namespace,
	}.Start()

	klog.InfoS("Listening", "address", serverAddr)
	utils.Check(http.ListenAndServe(serverAddr, RPC))
}
