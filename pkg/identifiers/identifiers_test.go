package identifiers

import (
	"reflect"
	"testing"
)

func TestMustParse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want Identifier
	}{
		{"basic1", args{"urn:publicid:IDN+plc:princeton+authority+sa"}, Identifier{
			Authorities:  []string{"plc", "princeton"},
			ResourceType: "authority",
			ResourceName: "sa",
		}},
		{"basic2", args{"urn:publicid:IDN+gcf:gpo:gpolab+user+joe"}, Identifier{
			Authorities:  []string{"gcf", "gpo", "gpolab"},
			ResourceType: "user",
			ResourceName: "joe",
		}},
		{"basic3", args{"urn:publicid:IDN+gcf:gpo:gpolab+node+switch+1+port+2"}, Identifier{
			Authorities:  []string{"gcf", "gpo", "gpolab"},
			ResourceType: "node",
			ResourceName: "switch+1+port+2",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustParse(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustParse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifier_Copy(t *testing.T) {
	type fields struct {
		Authorities  []string
		ResourceType string
		ResourceName string
	}
	type args struct {
		resourceType string
		resourceName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Identifier
	}{
		{
			"basic1",
			fields{
				Authorities:  []string{"gcf", "gpo", "gpolab"},
				ResourceType: "node",
				ResourceName: "switch+1+port+2",
			},
			args{"authority", "test"},
			Identifier{
				Authorities:  []string{"gcf", "gpo", "gpolab"},
				ResourceType: "authority",
				ResourceName: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Identifier{
				Authorities:  tt.fields.Authorities,
				ResourceType: tt.fields.ResourceType,
				ResourceName: tt.fields.ResourceName,
			}
			if got := v.Copy(tt.args.resourceType, tt.args.resourceName); !reflect.DeepEqual(
				got,
				tt.want,
			) {
				t.Errorf("Copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifier_URN(t *testing.T) {
	type fields struct {
		Authorities  []string
		ResourceType string
		ResourceName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"basic1",
			fields{
				Authorities:  []string{"gcf", "gpo", "gpolab"},
				ResourceType: "node",
				ResourceName: "switch+1+port+2",
			},
			"urn:publicid:IDN+gcf:gpo:gpolab+node+switch+1+port+2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Identifier{
				Authorities:  tt.fields.Authorities,
				ResourceType: tt.fields.ResourceType,
				ResourceName: tt.fields.ResourceName,
			}
			if got := v.URN(); got != tt.want {
				t.Errorf("URN() = %v, want %v", got, tt.want)
			}
		})
	}
}