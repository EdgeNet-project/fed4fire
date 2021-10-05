package rspec

import "encoding/xml"

type Rspec struct {
	XMLName xml.Name `xml:"rspec"`
	Type    string   `xml:"type,attr"`
	Nodes   []Node   `xml:"node"`
}

type Node struct {
	XMLName            xml.Name     `xml:"node"`
	ClientID           string       `xml:"client_id,attr,omitempty"`
	ComponentID        string       `xml:"component_id,attr"`
	ComponentManagerID string       `xml:"component_manager_id,attr"`
	ComponentName      string       `xml:"component_name,attr"`
	Exclusive          bool         `xml:"exclusive,attr"`
	HardwareType       HardwareType `xml:"hardware_type"`
	SliverTypes        []SliverType `xml:"sliver_type"`
	Available          Available    `xml:"available"`
	Location           Location     `xml:"location"`
}

type DiskImage struct {
	XMLName xml.Name `xml:"disk_image"`
	Name    string   `xml:"name,attr"`
}

type HardwareType struct {
	XMLName xml.Name `xml:"hardware_type"`
	Name    string   `xml:"name"`
}

type SliverType struct {
	XMLName    xml.Name    `xml:"sliver_type"`
	Name       string      `xml:"name,attr"`
	DiskImages []DiskImage `xml:"disk_image"`
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
