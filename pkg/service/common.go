package service

type Code struct {
	Code int `xml:"geni_code"`
}

type Credential struct {
	Geni_type    string `xml:"geni_type"`
	Geni_version int    `xml:"geni_version"`
	Geni_value   string `xml:"geni_value"`
}

type Options struct {
	// TODO: Fix lib to use tags.
	//Available    bool         `xml:"geni_available"`
	Geni_compressed    bool `xml:"geni_compressed"`
	Geni_rspec_version struct {
		Type    string
		Version string
	}
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
