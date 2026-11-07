package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{".id", "ID"},
		{"name", "Name"},
		{"local-address", "LocalAddress"},
		{"remote-address", "RemoteAddress"},
		{"caller-id", "CallerID"},
		{"dns-server", "DNSServer"},
		{"wins-server", "WINSServer"},
		{"change-tcp-mss", "ChangeTCPMSS"},
		{"use-upnp", "UseUPnP"},
		{"use-mpls", "UseMPLS"},
		{"limit-bytes-in", "LimitBytesIn"},
		{"limit-bytes-out", "LimitBytesOut"},
		{"use-radius", "UseRADIUS"},
		{"on-up", "OnUp"},
		{"on-down", "OnDown"},
		{"ipv6-routes", "IPv6Routes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMapRosTypeToGo(t *testing.T) {
	tests := []struct {
		rosName          string
		rosType          string
		isOptionalForSet bool
		want             string
	}{
		{"disabled", "bool", false, "bool"},
		{"disabled", "bool", true, "*bool"},
		{"limit-bytes-in", "num", false, "int64"},
		{"limit-bytes-in", "num", true, "*int64"},
		{"mtu", "num", false, "int"},
		{"mtu", "num", true, "*int"},
		{"local-address", "ip", false, "string"},
		{"name", "string", false, "string"},
	}

	for _, tt := range tests {
		t.Run(tt.rosName+"_"+tt.rosType, func(t *testing.T) {
			got := MapRosTypeToGo(tt.rosName, tt.rosType, tt.isOptionalForSet)
			if got != tt.want {
				t.Errorf("MapRosTypeToGo(%q, %q, %v) = %q, want %q", tt.rosName, tt.rosType, tt.isOptionalForSet, got, tt.want)
			}
		})
	}
}

func TestGenerateSource_PPP(t *testing.T) {
	schemas := getEmbeddedPPPSchema()
	opts := GeneratorOptions{
		PackageName: "ppp",
		RootMenu:    "ppp",
	}

	sourceBytes, err := GenerateSource(schemas, opts)
	if err != nil {
		t.Fatalf("GenerateSource returned error: %v", err)
	}

	source := string(sourceBytes)

	// Verify valid Go syntax by parsing AST
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "ppp.go", sourceBytes, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("Generated code failed Go parser:\n%v\n\nSource:\n%s", parseErr, source)
	}

	// Verify expected structs exist
	expectedStructs := []string{
		"type PPPSecret struct",
		"type PPPSecretAddRequest struct",
		"type PPPSecretSetRequest struct",
		"type PPPProfile struct",
		"type PPPProfileAddRequest struct",
		"type PPPProfileSetRequest struct",
		"type PPPActive struct",
		"type PPPAAA struct",
		"type PPPAAASetRequest struct",
	}

	for _, exp := range expectedStructs {
		if !strings.Contains(source, exp) {
			t.Errorf("Expected generated code to contain %q, but was missing", exp)
		}
	}

	// Verify JSON and ROS tags
	if !strings.Contains(source, `json:"local-address,omitempty" ros:"local-address"`) {
		t.Errorf("Expected json/ros tags for local-address")
	}
	if !strings.Contains(source, `json:"caller-id,omitempty" ros:"caller-id"`) {
		t.Errorf("Expected json/ros tags for caller-id")
	}
}
