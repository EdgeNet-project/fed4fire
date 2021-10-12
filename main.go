// This package implements the Fed4Fire Aggregate Manager API for EdgeNet/Kubernetes.
//
// Specifically, it implements the GENI AM API v3 as specified in
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3.
package main

import (
	"flag"
	"fmt"
	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"github.com/EdgeNet-project/fed4fire/pkg/service"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"github.com/gorilla/rpc"
	"github.com/maxmouchet/gorilla-xmlrpc/xml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strings"
)

var showHelp bool
var authorityName string
var containerImages utils.ArrayFlags
var containerCpuLimit string
var containerMemoryLimit string
var namespaceCpuLimit string
var namespaceMemoryLimit string
var kubeconfigFile string
var parentNamespace string
var serverAddr string
var trustedRootCerts utils.ArrayFlags

func beforeFunc(i *rpc.RequestInfo) {
	klog.InfoS(
		"Received XML-RPC request",
		"proto", i.Request.Proto,
		"method", i.Request.Method,
		"uri", i.Request.RequestURI,
		"rpc-method", i.Method,
		"user-agent", i.Request.UserAgent(),
	)
}

func main() {
	klog.InitFlags(nil)
	flag.BoolVar(&showHelp, "help", false, "show this message")
	flag.StringVar(&authorityName, "authorityName", "example.org", "authority name to use in URNs")
	flag.Var(&containerImages, "containerImage", "name:image of a container image that can be deployed; can be specified multiple times")
	flag.StringVar(&containerCpuLimit, "containerCpuLimit", "2", "maximum amount of CPU that can be used by a container")
	flag.StringVar(&containerMemoryLimit, "containerMemoryLimit", "2Gi", "maximum amount of memory that can be used by a container")
	flag.StringVar(&namespaceCpuLimit, "namespaceCpuLimit", "8", "maximum amount of CPU that can be used by a slice subnamespace")
	flag.StringVar(&namespaceMemoryLimit, "namespaceMemoryLimit", "8Gi", "maximum amount of memory that can be used by a slice subnamespace")
	flag.StringVar(&kubeconfigFile, "kubeconfig", "", "path to the kubeconfig file used to communicate with the Kubernetes API")
	flag.StringVar(&parentNamespace, "parentNamespace", "", "kubernetes namespaces in which to create slice subnamespaces")
	flag.StringVar(&serverAddr, "serverAddr", "localhost:9443", "host:port on which to listen")
	flag.Var(&trustedRootCerts, "trustedRootCert", "path to a trusted root certificate for authenticating user; can be specified multiple times")
	flag.Parse()

	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	utils.Check(err)

	edgeclient, err := versioned.NewForConfig(config)
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

	s := &service.Service{
		AbsoluteURL:          fmt.Sprintf("https://%s", serverAddr),
		AuthorityIdentifier:  authorityIdentifier,
		ContainerImages:      containerImages_,
		ContainerCpuLimit:    containerCpuLimit,
		ContainerMemoryLimit: containerMemoryLimit,
		NamespaceCpuLimit:    namespaceCpuLimit,
		NamespaceMemoryLimit: namespaceMemoryLimit,
		ParentNamespace:      parentNamespace,
		EdgenetClient:        edgeclient,
		KubernetesClient:     kubeclient,
	}

	xmlrpcCodec := xml.NewCodec()
	xmlrpcCodec.SetPrefix("Service.")

	RPC := rpc.NewServer()
	RPC.RegisterBeforeFunc(beforeFunc)
	RPC.RegisterCodec(xmlrpcCodec, "text/xml")
	utils.Check(RPC.RegisterService(s, ""))

	klog.InfoS("Listening", "address", serverAddr)
	utils.Check(http.ListenAndServe(serverAddr, RPC))
}
