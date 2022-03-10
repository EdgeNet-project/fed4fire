// Package identifiers implements GENI API Identifiers.
// https://groups.geni.net/geni/wiki/GeniApiIdentifiers
package identifiers

import (
	"fmt"
	"strings"

	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"github.com/oriser/regroup"
)

const (
	ResourceTypeAuthority = "authority"
	ResourceTypeImage     = "image"
	ResourceTypeNode      = "node"
	ResourceTypeSlice     = "slice"
	ResourceTypeSliver    = "sliver"
)

// `urn:publicid:IDN+toplevelauthority[:sub-authority]*\+resource-type\+resource-name`
var re = regroup.MustCompile(
	`urn:publicid:IDN\+(?P<authorities>.+?)\+(?P<resource_type>\w+)\+(?P<resource_name>[\w\+]+)`,
)

type Identifier struct {
	Authorities  []string
	ResourceType string
	ResourceName string
}

func (v Identifier) Copy(resourceType string, resourceName string) Identifier {
	authorities := make([]string, len(v.Authorities))
	copy(authorities, v.Authorities)
	return Identifier{
		Authorities:  authorities,
		ResourceType: resourceType,
		ResourceName: resourceName,
	}
}

func (v Identifier) Equal(vp Identifier) bool {
	return v.URN() == vp.URN()
}

func (v Identifier) URN() string {
	return fmt.Sprintf(
		"urn:publicid:IDN+%s+%s+%s",
		strings.Join(v.Authorities, ":"),
		v.ResourceType,
		v.ResourceName,
	)
}

func MustParse(s string) Identifier {
	identifier, err := Parse(s)
	utils.Check(err)
	return *identifier
}

func Parse(s string) (*Identifier, error) {
	matches, err := re.Groups(s)
	if err != nil {
		return nil, err
	}
	identifier := &Identifier{
		Authorities:  strings.Split(matches["authorities"], ":"),
		ResourceType: matches["resource_type"],
		ResourceName: matches["resource_name"],
	}
	return identifier, nil
}

func ParseMultiple(strings []string) ([]Identifier, error) {
	identifiers := make([]Identifier, len(strings))
	for i, s := range strings {
		identifier, err := Parse(s)
		if err != nil {
			return nil, err
		}
		identifiers[i] = *identifier
	}
	return identifiers, nil
}
