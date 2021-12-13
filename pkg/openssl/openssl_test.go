package openssl

import (
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
)

func TestVerify(t *testing.T) {
	parentCert, parentKey := utils.CreateCertificate("parent", "", "", nil, nil)
	childCert, _ := utils.CreateCertificate("child", "", "", parentCert, parentKey)
	err := Verify([][]byte{parentCert}, [][]byte{childCert})
	if err != nil {
		t.Errorf("Verify() = %s; want nil", err)
	}
	err = Verify([][]byte{}, [][]byte{childCert})
	if err == nil {
		t.Errorf("Verify() = nil; want error")
	}
}
