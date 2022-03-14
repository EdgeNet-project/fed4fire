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

// TODO
// Some things to take into account in request RSpecs:
// - Each node will have exactly one sliver_type in a request.
// - Each sliver_type will have zero or one disk_image elements.
//   If your testbed requires disk_image or does not support it,
//   it should handle bad requests RSpecs with the correct error.
// - The exclusive element is specified for each node in the request.
//   Your testbed should check if the specified value (in combination with the sliver_type) is supported,
//   and return the correct error if not.
// - The request RSpec might contain links that have a component_manager element that matches your AM.
//   If your AM does not support links, it should return the correct error.
// https://doc.fed4fire.eu/testbed_owner/rspec.html#request-rspec

// Some information will be in a request RSpec, that needs to be ignored and copied to the manifest RSpec unaltered.
// This is important to do correctly.
// - A request RSpec can contain nodes that have a component_manager_id set to a different AM.
//   You need to ignore these nodes, and copy them to the manifest RSpec unaltered.
// - A request RSpec can contain links that do not have a component_manager matching your AM
//   (links have multiple component_manager_id elements!).
//   You need to ignore these links, and copy them to the manifest RSpec unaltered.
// - A request RSpec can contain XML extensions in nodes, links, services, or directly in the rspec element.
//   Some of these the AM will not know.
//   It has to ignore these, and preferably also pass them unaltered to the manifest RSpec.
// https://doc.fed4fire.eu/testbed_owner/rspec.html#request-rspec

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

	// TODO: Implement RSpec passthroughs + arch selection + node selection.
	requestRspec := rspec.Rspec{}
	err = xml.Unmarshal([]byte(html.UnescapeString(args.Rspec)), &requestRspec)
	if err != nil {
		return reply.SetAndLogError(err, "Failed to deserialize rspec")
	}

	returnRspec := rspec.Rspec{Type: rspec.RspecTypeRequest}

	for i, node := range requestRspec.Nodes {
		sliverName := naming.SliverName(sliceIdentifier.URN(), node.ClientID)
		diskImage := s.ContainerImages[utils.Keys(s.ContainerImages)[0]]
		if len(node.SliverTypes) > 0 && len(node.SliverTypes[0].DiskImages) > 0 {
			if image, ok := s.ContainerImages[node.SliverTypes[0].DiskImages[0].Name]; ok {
				diskImage = image
			}
		}
		// TODO: Make sure the best effort tests pass.
		//identifier, err := identifiers.Parse(sliverType.DiskImages[0].Name)
		//if err != nil {
		//	return "", err
		//}
		//if image, ok := containerImages[identifier.ResourceName]; ok {
		//	return image, nil
		//} else {
		//	return "", fmt.Errorf("invalid image name")
		//}
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
				SliceURN: sliceIdentifier.URN(),
				UserURN:  userIdentifier.URN(),
				Expires:  metav1.NewTime(time.Now().Add(24 * time.Hour)),
				ClientID: node.ClientID,
				Image:    diskImage,
			},
		}
		sliver, err = s.Slivers().Create(r.Context(), sliver, metav1.CreateOptions{})
		if err != nil {
			return reply.SetAndLogError(err, "Failed to create sliver")
		}
		sliver.Status.AllocationStatus = constants.GeniStateAllocated
		sliver.Status.OperationalStatus = constants.GeniStateNotReady
		_, err = s.Slivers().UpdateStatus(r.Context(), sliver, metav1.UpdateOptions{})
		if err != nil {
			return reply.SetAndLogError(err, "Failed to update sliver status")
		}
		reply.Data.Value.Slivers = append(reply.Data.Value.Slivers, NewSliver(*sliver))
		returnRspec.Nodes = append(returnRspec.Nodes, requestRspec.Nodes[i])
	}

	xml_, err := xml.Marshal(returnRspec)
	if err != nil {
		return reply.SetAndLogError(err, "Failed to serialize response")
	}
	reply.Data.Value.Rspec = string(xml_)
	reply.Data.Code.Code = constants.GeniCodeSuccess
	return nil
}
