package cert

import (
	"testing"
)

func TestCertificateCommandsAndParsers(t *testing.T) {
	printCmd := NewPrintCertificatesCommand()
	if printCmd.Raw != "/certificate/print" || printCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid cert print: %v", printCmd)
	}

	addCmd := NewAddCertificateCommand("web-cert", "router.lan", 365)
	if addCmd.Raw != "/certificate/add" || addCmd.Args["days-valid"] != "365" {
		t.Errorf("unexpected add cert: %v", addCmd)
	}

	signCmd := NewSignCertificateCommand("*1", "MyCA")
	if signCmd.Raw != "/certificate/sign" || signCmd.Args["ca"] != "MyCA" {
		t.Errorf("unexpected sign cert: %v", signCmd)
	}

	certs := ParseCertificates([]map[string]string{
		{".id": "*1", "name": "web-cert", "common-name": "router.lan", "trusted": "true"},
	})
	if len(certs) != 1 || certs[0].Name != "web-cert" || !certs[0].Trusted {
		t.Errorf("unexpected certs parse: %+v", certs)
	}
}
