package rspec

import "encoding/xml"

//<rspec xmlns="http://www.geni.net/resources/rspec/3" type="request" xmlns:jfed="http://jfed.iminds.be/rspec/ext/jfed/1">
//    <node client_id="vm-one" component_manager_id="urn:publicid:IDN+example.com+authority+am" exclusive="false">
//        <sliver_type name="xen-vm">
//            <disk_image name="urn:publicid:IDN+example.con+image+ubuntu16"/>
//        </sliver_type>
//        <location xmlns="http://jfed.iminds.be/rspec/ext/jfed/1" x="100.0" y="100.0"/>
//    </node>
//    <node client_id="vm-two" component_manager_id="urn:publicid:IDN+example.com+authority+am" exclusive="false">
//        <sliver_type name="xen-vm">
//            <disk_image name="urn:publicid:IDN+example.con+image+ubuntu16"/>
//        </sliver_type>
//        <jfed:location x="100.0" y="100.0"/>
//    </node>
//</rspec>

type Request struct {
	XMLName xml.Name `xml:"rspec"`
	Type    string   `xml:"type,attr"`
	Nodes   []Node
}