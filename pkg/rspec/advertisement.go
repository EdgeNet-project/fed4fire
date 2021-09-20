package rspec

import "encoding/xml"

//<node xmlns:emulab="http://www.protogeni.net/resources/rspec/ext/emulab/1" component_id="urn:publicid:IDN+example.com+node+vmhost1" component_manager_id="urn:publicid:IDN+example.com+authority+am" component_name="vmhost1" exclusive="false" >
//    <sliver_type name="xen-vm">
//        <disk_image name="urn:publicid:IDN+example.con+image+ubuntu14"/>
//        <disk_image name="urn:publicid:IDN+example.con+image+ubuntu16"/>
//    </sliver_type>
//    <hardware_type name="pc-vmhost">
//        <emulab:node_type type_slots="20"/>
//    </hardware_type>
//    <available now="true"/>
//    <location country="NU" latitude="0.0" longitude="0.0"/>4
//</node>

type DiskImage struct {
	XMLName xml.Name `xml:"disk_image"`
	Name    string   `xml:"name,attr"`
}

type HardwareType struct {
	XMLName xml.Name `xml:"hardware_type"`
	Name    string   `xml:"name"`
}

type SliverType struct {
	XMLName    xml.Name `xml:"sliver_type"`
	Name       string   `xml:"name,attr"`
	DiskImages []DiskImage
}

type Node struct {
	XMLName            xml.Name `xml:"node"`
	ComponentID        string   `xml:"component_id,attr"`
	ComponentManagerID string   `xml:"component_manager_id,attr"`
	ComponentName      string   `xml:"component_name,attr"`
	Exclusive          bool     `xml:"exclusive,attr"`
	HardwareType       HardwareType
	SliverTypes        []SliverType
}

type Advertisement struct {
	XMLName xml.Name `xml:"rspec"`
	Type    string   `xml:"type,attr"`
	Nodes   []Node
}
