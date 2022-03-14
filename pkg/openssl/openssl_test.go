package openssl

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
)

func TestVerify(t *testing.T) {
	parentCert, parentKey := utils.CreateCertificate("parent", "", "", nil, nil)
	childCert, _ := utils.CreateCertificate("child", "", "", parentCert, parentKey)
	assert.Nil(t, Verify([][]byte{parentCert}, [][]byte{childCert}))
	assert.NotNil(t, Verify([][]byte{}, [][]byte{childCert}))
}
