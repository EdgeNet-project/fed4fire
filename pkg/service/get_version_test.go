package service

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	s := Service{}
	args := &GetVersionArgs{}
	reply := &GetVersionReply{}
	err := s.GetVersion(nil, args, reply)
	if err != nil {
		t.Errorf("GetVersion() = %v; want nil", err)
	}
	if got := reply.Data.Code.Code; got != geniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, geniCodeSuccess)
	}
	if got := len(reply.Data.Value.AdRspecVersions); got != 1 {
		t.Errorf("len(AdRspecVersions) = %d; want 1", got)
	}
	if got := len(reply.Data.Value.RequestRspecVersions); got != 1 {
		t.Errorf("len(RequestRspecVersions) = %d; want 1", got)
	}
	if got := len(reply.Data.Value.CredentialTypes); got != 1 {
		t.Errorf("len(CredentialTypes) = %d; want 1", got)
	}
}
