package service

import "net/http"

type PerformOperationalActionArgs struct {
	URNs        []string
	Credentials []Credential
	Action      string
	Options     string // Options
}

type PerformOperationalActionReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

// PerformOperationalAction performs the named operational action on the named slivers,
// possibly changing the geni_operational_status of the named slivers, e.g. 'start' a VM.
// For valid operations and expected states, consult the state diagram advertised in the aggregate's advertisement RSpec.
func (s *Service) PerformOperationalAction(r *http.Request, args *PerformOperationalActionArgs, reply *PerformOperationalActionReply) error {
	return nil
}
