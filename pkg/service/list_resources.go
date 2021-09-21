package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
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
	fmt.Println(args)

	// TODO: Proper error structs.
	nodes, err := s.KubernetesClient.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return err
	}

	v := rspec.Advertisement{
		Type: "advertisement",
	}

	for _, node := range nodes.Items {
		nodeArch := node.Labels["kubernetes.io/arch"]
		nodeName := node.Name
		node_ := rspec.Node{
			ComponentID:        fmt.Sprintf("urn:publicid:IDN+edge-net.org+node+%s", nodeName),
			ComponentManagerID: s.URN,
			ComponentName:      nodeName,
			HardwareType: rspec.HardwareType{
				Name: fmt.Sprintf("kubernetes-%s", nodeArch),
			},
			SliverTypes: []rspec.SliverType{
				{
					Name: "container",
					DiskImages: []rspec.DiskImage{
						{Name: "urn:publicid:IDN+edge-net.org+image+ubuntu2004"},
					},
				},
			},
		}
		v.Nodes = append(v.Nodes, node_)
	}

	reply.Data.Code.Code = geni_code_success
	xml_, err := xml.Marshal(v)
	if err != nil {
		fmt.Println(err)
		return err
	}
	reply.Data.Value = string(xml_)
	return nil
}