package service

import "net/http"

type DescribeArgs struct {
	URNs        []string
	Credentials []Credential
	Options     Options
}

type DescribeReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

// Describe retrieves a manifest RSpec describing the resources contained by the named entities,
// e.g. a single slice or a set of the slivers in a slice.
// This listing and description should be sufficiently descriptive to allow experimenters to use the resources.
func (s *Service) Describe(r *http.Request, args *DescribeArgs, reply *DescribeReply) error {
	return nil
}
