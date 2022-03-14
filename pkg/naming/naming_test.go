package naming

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	"k8s.io/apimachinery/pkg/util/validation"
)

var testSliceIdentifier = identifiers.MustParse("urn:publicid:IDN+example.org+slice+test")

func TestSliceHash(t *testing.T) {
	h := SliceHash(testSliceIdentifier.URN())
	errs := validation.IsValidLabelValue(h)
	assert.Len(t, errs, 0)
}

func TestSliverName(t *testing.T) {
	h := SliverName(testSliceIdentifier.URN(), "Client$Id&*")
	errs := validation.IsValidLabelValue(h)
	assert.Len(t, errs, 0)
}
