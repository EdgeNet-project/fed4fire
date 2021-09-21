package service

import (
	"net/http"
)

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
			URN                  string           `xml:"urn"`
			API                  int              `xml:"geni_api"`
			APIVersions          APIVersions      `xml:"geni_api_versions"`
			RequestRspecVersions []RspecVersion   `xml:"geni_request_rspec_versions"`
			AdRspecVersions      []RspecVersion   `xml:"geni_ad_rspec_versions"`
			CredentialTypes      []CredentialType `xml:"geni_credential_types"`
			SingleAllocation     int              `xml:"geni_single_allocation"`
			Allocate             string           `xml:"geni_allocate"`
		} `xml:"value"`
	}
}

func (s *Service) GetVersion(r *http.Request, args *GetVersionArgs, reply *GetVersionReply) error {
	reply.Data.API = 3
	reply.Data.Code.Code = geni_code_success
	reply.Data.Value.URN = s.URN
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
	//// TODO: _enum_ for geni_sfa, geni_many...?
	reply.Data.Value.CredentialTypes = []CredentialType{
		{
			Type:    "geni_sfa",
			Version: "3",
		},
	}
	reply.Data.Value.SingleAllocation = 0
	reply.Data.Value.Allocate = "geni_many"
	return nil
}