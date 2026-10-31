package cert

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// NewPrintCertificatesCommand queries certificates with strict proplist.
func NewPrintCertificatesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "common-name", "fingerprint", "invalid-before", "invalid-after", "expires-after", "ca", "trusted", "issued")
	return command.Command{
		Raw:  "/certificate/print",
		Args: args,
	}
}

// NewAddCertificateCommand generates a certificate request/template.
func NewAddCertificateCommand(name, commonName string, daysValid int) command.Command {
	args := map[string]string{
		"name":        name,
		"common-name": commonName,
	}
	if daysValid > 0 {
		args["days-valid"] = strconv.Itoa(daysValid)
	}
	return command.Command{
		Raw:  "/certificate/add",
		Args: args,
	}
}

// NewSignCertificateCommand signs a certificate template with a CA.
func NewSignCertificateCommand(id, ca string) command.Command {
	args := map[string]string{
		"number": id,
	}
	if ca != "" {
		args["ca"] = ca
	}
	return command.Command{
		Raw:  "/certificate/sign",
		Args: args,
	}
}

// NewRemoveCertificateCommand deletes a certificate by ID.
func NewRemoveCertificateCommand(id string) command.Command {
	return command.Command{
		Raw: "/certificate/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}
