package port

import (
	domainPPP "github.com/quixiq/polyglot/internal/domain/ppp"
)

// PPPActiveSession alias to domain struct.
type PPPActiveSession = domainPPP.ActiveSession

// PPPoESecret alias to domain struct.
type PPPoESecret = domainPPP.PPPoESecret

// PPPProfile alias to domain struct.
type PPPProfile = domainPPP.Profile

// PPPoESecretParams holds parameters for creating or updating a PPPoE secret.
type PPPoESecretParams struct {
	Name          string
	Password      string
	Profile       string
	Service       string
	LocalAddress  string
	RemoteAddress string
	Comment       string
	Disabled      bool
	CallerID      string
}

// PPPProfileParams holds parameters for creating or updating a PPP profile.
type PPPProfileParams struct {
	Name           string
	RateLimit      string
	LocalAddress   string
	RemoteAddress  string
	DNSServer      string
	ParentQueue    string
	AddressList    string
	Comment        string
	SessionTimeout string
	IdleTimeout    string
	SharedUsers    string
	OnlyOne        string
	UseMPLS        string
	UseCompression string
	UseEncryption  string
	ChangeTCPMSS   string
	BridgeLearning string
	OnUp           string
	OnDown         string
}
