package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

type ListResourcesArgs struct {
	Credentials []Credential
	Options     Options
}

type ListResourcesReply struct {
	Data struct {
		Code  Code   `xml:"code"`
		Value string `xml:"value"`
	}
}

func (s *Service) ListResources(r *http.Request, args *ListResourcesArgs, reply *ListResourcesReply) error {
	fmt.Println(args.Options)
	// TODO: Proper error structs.
	nodes, err := s.KubernetesClient.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return err
	}

	v := rspec.Rspec{
		Type: "advertisement",
	}

	for _, node := range nodes.Items {
		nodeArch := node.Labels["kubernetes.io/arch"]
		nodeCountry := node.Labels["edge-net.io/country-iso"]
		nodeLatitude := node.Labels["edge-net.io/lat"][1:]
		nodeLongitude := node.Labels["edge-net.io/lon"][1:]
		nodeName := node.Name
		nodeIsReady := true
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				nodeIsReady = false
			}
		}
		node_ := rspec.Node{
			ComponentID:        s.URN("node", nodeName),
			ComponentManagerID: s.URN("authority", "am"),
			ComponentName:      nodeName,
			Available:          rspec.Available{Now: nodeIsReady},
			Location: rspec.Location{
				Country:   nodeCountry,
				Latitude:  nodeLatitude,
				Longitude: nodeLongitude,
			},
			HardwareType: rspec.HardwareType{
				Name: fmt.Sprintf("kubernetes-%s", nodeArch),
			},
			SliverTypes: []rspec.SliverType{
				{
					Name: "container",
					DiskImages: []rspec.DiskImage{
						{Name: s.URN("image", "ubuntu2004")},
					},
				},
			},
		}
		if !(args.Options.Available && !node_.Available.Now) {
			v.Nodes = append(v.Nodes, node_)
		}
	}

	reply.Data.Code.Code = geniCodeSuccess
	xml_, err := xml.Marshal(v)
	if err != nil {
		return err
	}

	if args.Options.Compressed {
		reply.Data.Value = utils.ZlibBase64(xml_)
	} else {
		reply.Data.Value = string(xml_)
	}

	return nil
}
