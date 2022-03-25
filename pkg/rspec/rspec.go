package rspec

import "encoding/xml"

const (
	RspecTypeAdvertisement = "advertisement"
	RspecTypeManifest      = "manifest"
	RspecTypeRequest       = "request"
)

const (
	RspecLoginAuthenticationSSH = "ssh-keys"
)

type Rspec struct {
	XMLName xml.Name `xml:"http://www.geni.net/resources/rspec/3 rspec"`
	Type    string   `xml:"type,attr"`
	Nodes   []Node   `xml:"node"`
}

type Node struct {
	XMLName            xml.Name      `xml:"node"`
	ClientID           string        `xml:"client_id,attr,omitempty"`
	ComponentID        string        `xml:"component_id,attr,omitempty"`
	ComponentManagerID string        `xml:"component_manager_id,attr,omitempty"`
	ComponentName      string        `xml:"component_name,attr,omitempty"`
	SliverID           string        `xml:"sliver_id,attr,omitempty"`
	Exclusive          bool          `xml:"exclusive,attr"`
	HardwareType       *HardwareType `xml:"hardware_type,omitempty"`
	SliverType         SliverType    `xml:"sliver_type"`
	Services           *Services     `xml:"services,omitempty"`
	Available          *Available    `xml:"available,omitempty"`
	Location           *Location     `xml:"location,omitempty"`
}

type DiskImage struct {
	XMLName xml.Name `xml:"disk_image"`
	Name    string   `xml:"name,attr"`
}

type HardwareType struct {
	XMLName xml.Name `xml:"hardware_type"`
	Name    string   `xml:"name,attr"`
}

type SliverType struct {
	XMLName    xml.Name    `xml:"sliver_type"`
	Name       string      `xml:"name,attr"`
	DiskImages []DiskImage `xml:"disk_image"`
}

type Services struct {
	XMLName xml.Name `xml:"services"`
	Logins  []Login  `xml:"login"`
}

type Login struct {
	XMLName        xml.Name `xml:"login"`
	Authentication string   `xml:"authentication,attr"`
	Hostname       string   `xml:"hostname,attr"`
	Port           int      `xml:"port,attr"`
	Username       string   `xml:"username,attr"`
}

type Available struct {
	XMLName xml.Name `xml:"available"`
	Now     bool     `xml:"now,attr"`
}

type Location struct {
	XMLName   xml.Name `xml:"location"`
	Country   string   `xml:"country,attr"`
	Latitude  string   `xml:"latitude,attr"`
	Longitude string   `xml:"longitude,attr"`
}
