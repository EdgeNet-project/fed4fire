package service

import "net/http"

type StatusArgs struct {
	URNs        []string
	Credentials []Credential
	Options     string // Options
}

type StatusReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

// Status gets the status of a sliver or slivers belonging to a single slice at the given aggregate.
// Status may include other dynamic reservation or instantiation information as required by the resource type and aggregate.
// This method is used to provide updates on the state of the resources after the completion of Provision,
// which began to asynchronously provision the resources. This should be relatively dynamic data,
// not descriptive data as returned in the manifest RSpec.
func (s *Service) Status(r *http.Request, args *StatusArgs, reply *StatusReply) error {
	return nil
}
