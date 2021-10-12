package service

import (
	"testing"
)

func TestAllocate_NoCredentials(t *testing.T) {
	s := testService()
	r := testRequest()
	args := &AllocateArgs{SliceURN: testSliceUrn}
	reply := &AllocateReply{}
	err := s.Allocate(r, args, reply)
	if err != nil {
		t.Errorf("Allocate() = %v; want nil", err)
	}
	got := reply.Data.Code.Code
	if got != geniCodeError {
		t.Errorf("Code = %d; want %d", got, geniCodeError)
	}
}
