package service

import (
	"encoding/xml"
	"fmt"
	v1 "github.com/EdgeNet-project/fed4fire/pkg/apis/fed4fire/v1"
	"github.com/EdgeNet-project/fed4fire/pkg/constants"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"html"
	"net/http"
	"time"

	"github.com/EdgeNet-project/fed4fire/pkg/naming"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"github.com/EdgeNet-project/fed4fire/pkg/rspec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type AllocateArgs struct {
	SliceURN    string
	Credentials []Credential
	Rspec       string
	Options     Options
}

type AllocateReply struct {
	Data struct {
		Code   Code   `xml:"code"`
		Output string `xml:"output"`
		Value  struct {
			Rspec   string   `xml:"geni_rspec"`
			Slivers []Sliver `xml:"geni_slivers"`
		} `xml:"value"`
	}
}

func (v *AllocateReply) SetAndLogError(err error, msg string, keysAndValues ...interface{}) error {
	klog.ErrorS(err, msg, keysAndValues...)
	v.Data.Code.Code = constants.GeniCodeError
	v.Data.Output = fmt.Sprintf("%s: %s", msg, err)
	return nil
}

// Allocate allocates resources as described in a request RSpec argument to a slice with the named URN.
// On success, one or more slivers are allocated, containing resources satisfying the request, and assigned to the given slice.
// This method returns a listing and description of the resources reserved for the slice by this operation, in the form of a manifest RSpec.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#Allocate
func (s *Service) Allocate(r *http.Request, args *AllocateArgs, reply *AllocateReply) error {
	userIdentifier, err := identifiers.Parse(r.Header.Get(constants.HttpHeaderUser))
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorBadIdentifier)
	}
	sliceIdentifier, err := identifiers.Parse(args.SliceURN)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorBadIdentifier)
	}
	_, err = FindCredential(
		*userIdentifier,
		sliceIdentifier,
		args.Credentials,
		s.TrustedCertificates,
	)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorBadCredentials)
	}

	requestRspec := rspec.Rspec{}
	err = xml.Unmarshal([]byte(html.UnescapeString(args.Rspec)), &requestRspec)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorDeserializeRspec)
	}

	returnRspec := rspec.Rspec{Type: rspec.RspecTypeRequest}

	for i, node := range requestRspec.Nodes {
		sliverName := naming.SliverName(sliceIdentifier.URN(), node.ClientID)
		// We're very lenient here: if there is no image specified, or
		// if a disk image is specified but does not exist, we use a default one.
		diskImage := s.ContainerImages[utils.Keys(s.ContainerImages)[0]]
		if len(node.SliverType.DiskImages) > 0 {
			if image, ok := s.ContainerImages[node.SliverType.DiskImages[0].Name]; ok {
				diskImage = image
			}
		}
		var requestedArch *string
		if node.HardwareType.Name != "" {
			requestedArch = &node.HardwareType.Name
		}
		var requestedNode *string
		if node.ComponentID != "" {
			componentId, err := identifiers.Parse(node.ComponentID)
			if err != nil {
				return reply.SetAndLogError(err, constants.ErrorBadIdentifier)
			}
			requestedNode = &componentId.ResourceName
		}
		labels := map[string]string{
			// We store the hash since the full URN would not be a valid label value;
			// this allows us to easily get all the resources belonging to a slice.
			constants.Fed4FireSliceHash:  naming.SliceHash(sliceIdentifier.URN()),
			constants.Fed4FireSliverName: sliverName,
		}
		sliver := &v1.Sliver{
			ObjectMeta: metav1.ObjectMeta{
				Name:   sliverName,
				Labels: labels,
			},
			Spec: v1.SliverSpec{
				URN: s.AuthorityIdentifier.Copy(identifiers.ResourceTypeSliver, sliverName).
					URN(),
				SliceURN:      sliceIdentifier.URN(),
				UserURN:       userIdentifier.URN(),
				Expires:       metav1.NewTime(time.Now().Add(24 * time.Hour)),
				ClientID:      node.ClientID,
				Image:         diskImage,
				RequestedArch: requestedArch,
				RequestedNode: requestedNode,
			},
		}
		sliver, err = s.Slivers().Create(r.Context(), sliver, metav1.CreateOptions{})
		if err != nil {
			return reply.SetAndLogError(err, constants.ErrorCreateResource)
		}
		allocationStatus, operationalStatus := s.GetSliverStatus(r.Context(), sliver.Name)
		reply.Data.Value.Slivers = append(
			reply.Data.Value.Slivers,
			NewSliver(*sliver, allocationStatus, operationalStatus),
		)
		returnRspec.Nodes = append(returnRspec.Nodes, requestRspec.Nodes[i])
	}

	xml_, err := xml.Marshal(returnRspec)
	if err != nil {
		return reply.SetAndLogError(err, constants.ErrorSerializeRspec)
	}
	reply.Data.Value.Rspec = string(xml_)
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
