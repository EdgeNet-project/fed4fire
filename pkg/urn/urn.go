package urn

import (
	"fmt"
	"github.com/EdgeNet-project/fed4fire/pkg/utils"
	"github.com/oriser/regroup"
	"strings"
)

// https://groups.geni.net/geni/wiki/GeniApiIdentifiers
// `urn:publicid:IDN+toplevelauthority[:sub-authority]*\+resource-type\+resource-name`
var re = regroup.MustCompile(
	`urn:publicid:IDN\+(?P<authorities>.+?)\+(?P<resource_type>\w+)\+(?P<resource_name>[\w\+]+)`,
)

type Identifier struct {
	Authorities  []string
	ResourceType string
	ResourceName string
}

func (v Identifier) String() string {
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
