package service

import "net/http"

type ShutdownArgs struct {
	SliceURN    string
	Credentials []Credential
	Options     Options
}

type ShutdownReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

// Shutdown performs an emergency shutdown on the slivers in the given slice at this aggregate.
// Resources should be taken offline, such that experimenter access (on both the control and data plane) is cut off.
// No further actions on the slivers in the given slice should be possible at this aggregate,
// until an un-specified operator action restores the slice's slivers (or deletes them).
// This operation is intended for operator use.
// The slivers are shut down but remain available for further forensics.
func (s *Service) Shutdown(r *http.Request, args *ShutdownArgs, reply *ShutdownReply) error {
	return nil
}
