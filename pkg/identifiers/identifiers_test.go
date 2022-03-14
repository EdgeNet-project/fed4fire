package identifiers

import (
	"reflect"
	"testing"
)

func TestMustParse(t *testing.T) {
	tests := []struct {
		urn  string
		want Identifier
	}{
		{
			"urn:publicid:IDN+plc:princeton+authority+sa",
			Identifier{
				Authorities:  []string{"plc", "princeton"},
				ResourceType: "authority",
				ResourceName: "sa",
			},
		},
		{
			"urn:publicid:IDN+gcf:gpo:gpolab+user+joe",
			Identifier{
				Authorities:  []string{"gcf", "gpo", "gpolab"},
				ResourceType: "user",
				ResourceName: "joe",
			},
		},
		{
			"urn:publicid:IDN+gcf:gpo:gpolab+node+switch+1+port+2",
			Identifier{
				Authorities:  []string{"gcf", "gpo", "gpolab"},
				ResourceType: "node",
				ResourceName: "switch+1+port+2",
			},
		},
		{

			"urn:publicid:IDN+edge-net.org+node+geni-us-tn-cb07.edge-net.io",
			Identifier{
				Authorities:  []string{"edge-net.org"},
				ResourceType: "node",
				ResourceName: "geni-us-tn-cb07.edge-net.io",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.urn, func(t *testing.T) {
			if got := MustParse(tt.urn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustParse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifier_Copy(t *testing.T) {
	id := Identifier{
		Authorities:  []string{"gcf", "gpo", "gpolab"},
		ResourceType: "node",
		ResourceName: "switch+1+port+2",
	}
	want := Identifier{
		Authorities:  []string{"gcf", "gpo", "gpolab"},
		ResourceType: "authority",
		ResourceName: "test",
	}
	if got := id.Copy("authority", "test"); !reflect.DeepEqual(got, want) {
		t.Errorf("Copy() = %v, want %v", got, want)
	}
}

func TestIdentifier_URN(t *testing.T) {
	id := Identifier{
		Authorities:  []string{"gcf", "gpo", "gpolab"},
		ResourceType: "node",
		ResourceName: "switch+1+port+2",
	}
	want := "urn:publicid:IDN+gcf:gpo:gpolab+node+switch+1+port+2"
	if got := id.URN(); got != want {
		t.Errorf("URN() = %v, want %v", got, want)
	}
}
