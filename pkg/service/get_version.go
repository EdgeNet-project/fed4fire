package service

import "net/http"

type APIVersions struct {
	Three string `xml:"3"`
}

type CredentialType struct {
	Type    string `xml:"geni_type"`
	Version string `xml:"geni_version"`
}

type RspecVersion struct {
	Type       string   `xml:"type"`
	Version    string   `xml:"version"`
	Schema     string   `xml:"schema"`
	Namespace  string   `xml:"namespace"`
	Extensions []string `xml:"extensions"`
}

type GetVersionArgs struct{}

type GetVersionReply struct {
	Data struct {
		API   int  `xml:"geni_api"`
		Code  Code `xml:"code"`
		Value struct {
			URN string `xml:"urn"`
			// Current version of this API.
			API int `xml:"geni_api"`
			// List of versions of the API supported by this aggregate.
			APIVersions APIVersions `xml:"geni_api_versions"`
			// List of request RSpec formats supported by this aggregate.
			RequestRspecVersions []RspecVersion `xml:"geni_request_rspec_versions"`
			// List of advertisement RSpec formats supported by this aggregate.
			AdRspecVersions []RspecVersion `xml:"geni_ad_rspec_versions"`
			// List of supported credential types and versions.
			CredentialTypes []CredentialType `xml:"geni_credential_types"`
			// When true (not default), and performing one of (Describe, Allocate, Renew, Provision, Delete),
			// such an AM requires you to include either the slice urn or the urn of all the slivers in the same state.
			// If you attempt to run one of those operations on just some slivers in a given state,
			// such an AM will return an error.
			SingleAllocation int `xml:"geni_single_allocation"`
			// Defines whether this AM allows adding slivers to slices at an AM.
			Allocate string `xml:"geni_allocate"`
		} `xml:"value"`
	}
}

// GetVersion returns static configuration information about this aggregate manager implementation,
// such as API and RSpec versions supported.
// https://groups.geni.net/geni/wiki/GAPI_AM_API_V3#GetVersion
func (s *Service) GetVersion(r *http.Request, args *GetVersionArgs, reply *GetVersionReply) error {
	reply.Data.API = 3
	reply.Data.Value.URN = s.AuthorityIdentifier.URN()
	reply.Data.Value.API = 3
	reply.Data.Value.APIVersions.Three = s.AbsoluteURL
	reply.Data.Value.RequestRspecVersions = []RspecVersion{
		{
			Type:      "GENI",
			Version:   "3",
			Schema:    "http://www.geni.net/resources/rspec/3/request.xsd",
			Namespace: "http://www.geni.net/resources/rspec/3",
		},
	}
	reply.Data.Value.AdRspecVersions = []RspecVersion{
		{
			Type:      "GENI",
			Version:   "3",
			Schema:    "http://www.geni.net/resources/rspec/3/ad.xsd",
			Namespace: "http://www.geni.net/resources/rspec/3",
		},
	}
	reply.Data.Value.CredentialTypes = []CredentialType{
		{
			Type:    geniCredentialTypeSfa,
			Version: "3",
		},
	}
	reply.Data.Value.SingleAllocation = 0
	reply.Data.Value.Allocate = geniAllocateMany
	reply.Data.Code.Code = geniCodeSuccess
	return nil
}
