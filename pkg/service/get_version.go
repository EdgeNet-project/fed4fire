package service

import (
	"log"
	"net/http"
)

type GetVersionArgs struct{}

type CredentialType struct {
	Type    string `xml:"geni_type"`
	Version int    `xml:"geni_version"`
}

type RspecVersion struct {
	Type       string   `xml:"type"`
	Version    int      `xml:"version"`
	Schema     string   `xml:"schema"`
	Namespace  string   `xml:"namespace"`
	Extensions []string `xml:"extensions"`
}

// TODO: Refactor structs
type GetVersionReply struct {
	Data struct {
		API  int `xml:"geni_api"`
		Code struct {
			Code int `xml:"geni_code"`
		} `xml:"code"`
		Value struct {
			URN         string `xml:"urn"`
			API         int    `xml:"geni_api"`
			APIVersions struct {
				Three string `xml:"3"`
			} `xml:"geni_api_versions"`
			RequestRspecVersions []RspecVersion   `xml:"geni_request_rspec_versions"`
			AdRspecVersions      []RspecVersion   `xml:"geni_ad_rspec_versions"`
			CredentialTypes      []CredentialType `xml:"geni_credential_types"`
			SingleAllocation     int              `xml:"geni_single_allocation"`
			Allocate             string           `xml:"geni_allocate"`
		} `xml:"value"`
	}
}

func (s *Service) GetVersion(r *http.Request, args *GetVersionArgs, reply *GetVersionReply) error {
	log.Println("GetVersion")
	reply.Data.API = 3
	reply.Data.Code.Code = 0
	reply.Data.Value.URN = "urn:publicid:IDN+edge-net.org+authority+am"
	reply.Data.Value.API = 3
	reply.Data.Value.APIVersions.Three = s.AbsoluteURL
	// TODO: Check the values below.
	// What are the required specs?
	// What are the default values for some fields?
	//reply.Data.Value.RequestRspecVersions = []RspecVersion{
	//	{
	//		Type:      "GENI",
	//		Version:   3,
	//		Schema:    "http://www.geni.net/resources/rspec/3/request.xsd",
	//		Namespace: "http://www.geni.net/resources/rspec/3",
	//	},
	//}
	//reply.Data.Value.AdRspecVersions = []RspecVersion{
	//	{
	//		Type:       "GENI",
	//		Version:    3,
	//		Schema:     "http://www.geni.net/resources/rspec/3/ad.xsd",
	//		Namespace:  "http://www.geni.net/resources/rspec/3",
	//		Extensions: nil,
	//	},
	//}
	//// TODO: _enum_ for geni_sfa, geni_many...?
	//reply.Data.Value.CredentialTypes = []CredentialType{
	//	{
	//		Type:    "geni_sfa",
	//		Version: 3,
	//	},
	//}
	reply.Data.Value.SingleAllocation = 0
	reply.Data.Value.Allocate = "geni_many"
	return nil
}
