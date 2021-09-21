package service

type Code struct {
	Code int `xml:"geni_code"`
}

type Credential struct {
	Type    string `xml:"geni_type"`
	Version int    `xml:"geni_version"`
	Value   string `xml:"geni_value"`
}

type Options struct {
	Available    bool `xml:"geni_available"`
	Cmpressed    bool `xml:"geni_compressed"`
	RspecVersion struct {
		Type    string
		Version string
	} `xml:"geni_rspec_version"`
}

const (
	geni_code_success                = 0
	geni_code_badargs                = 1
	geni_code_error                  = 2
	geni_code_forbidden              = 3
	geni_code_badversion             = 4
	geni_code_serverror              = 5
	geni_code_toobig                 = 6
	geni_code_refused                = 7
	geni_code_timedout               = 8
	geni_code_dberror                = 9
	geni_code_rpcerror               = 10
	geni_code_unavailable            = 11
	geni_code_searchfailed           = 12
	geni_code_unsupported            = 13
	geni_code_busy                   = 14
	geni_code_expired                = 15
	geni_code_inprogress             = 16
	geni_code_alreadyexists          = 17
	geni_code_vlan_unavailable       = 24
	geni_code_insufficient_bandwidth = 25
)
