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
	got := reply.Data.Code.Code
	if got != geniCodeSuccess {
		t.Errorf("Code = %d; want %d", got, geniCodeSuccess)
	}
	got = len(reply.Data.Value.AdRspecVersions)
	if got != 1 {
		t.Errorf("len(AdRspecVersions) = %d; want 1", got)
	}
	got = len(reply.Data.Value.RequestRspecVersions)
	if got != 1 {
		t.Errorf("len(RequestRspecVersions) = %d; want 1", got)
	}
	got = len(reply.Data.Value.CredentialTypes)
	if got != 1 {
		t.Errorf("len(CredentialTypes) = %d; want 1", got)
	}
}
