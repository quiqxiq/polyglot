package ppp

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
)

// toProtoPPPSecret converts a domain port.PPPoESecret to protobuf PPPSecret.
func toProtoPPPSecret(s port.PPPoESecret) *devicepb.PPPSecret {
	return &devicepb.PPPSecret{
		Id:            s.RosID,
		Name:          s.Name,
		Password:      "", // password is kept write-only/redacted in lists
		Profile:       s.Profile,
		Service:       s.Service,
		LocalAddress:  s.LocalAddress,
		RemoteAddress: s.RemoteAddress,
		Comment:       s.Comment,
		Disabled:      s.Disabled,
		LastLoggedOut: s.LastLoggedOut,
		CallerId:      s.CallerID,
	}
}

// toProtoPPPSecrets converts a slice of domain port.PPPoESecret to protobuf PPPSecrets.
func toProtoPPPSecrets(secrets []port.PPPoESecret) []*devicepb.PPPSecret {
	res := make([]*devicepb.PPPSecret, len(secrets))
	for i, s := range secrets {
		res[i] = toProtoPPPSecret(s)
	}
	return res
}

// toProtoPPPProfile converts a domain port.PPPProfile to protobuf PPPProfile.
func toProtoPPPProfile(p port.PPPProfile) *devicepb.PPPProfile {
	return &devicepb.PPPProfile{
		Id:             p.RosID,
		Name:           p.Name,
		RateLimit:      p.RateLimit,
		LocalAddress:   p.LocalAddress,
		RemoteAddress:  p.RemoteAddress,
		DnsServer:      p.DNSServer,
		ParentQueue:    p.ParentQueue,
		AddressList:    p.AddressList,
		Comment:        p.Comment,
		SharedUsers:    p.SharedUsers,
		OnlyOne:        p.OnlyOne,
		UseMpls:        p.UseMPLS,
		UseCompression: p.UseCompression,
		UseEncryption:  p.UseEncryption,
		ChangeTcpMss:   p.ChangeTCPMSS,
		BridgeLearning: p.BridgeLearning,
	}
}

// toProtoPPPProfiles converts a slice of domain port.PPPProfile to protobuf PPPProfiles.
func toProtoPPPProfiles(profiles []port.PPPProfile) []*devicepb.PPPProfile {
	res := make([]*devicepb.PPPProfile, len(profiles))
	for i, p := range profiles {
		res[i] = toProtoPPPProfile(p)
	}
	return res
}

// toProtoPPPActiveSession converts a domain port.PPPActiveSession to protobuf PPPActiveSession.
func toProtoPPPActiveSession(s port.PPPActiveSession) *devicepb.PPPActiveSession {
	return &devicepb.PPPActiveSession{
		Id:        s.RosID,
		Name:      s.Name,
		Service:   s.Service,
		CallerId:  s.CallerID,
		Address:   s.Address,
		Encoding:  s.Encoding,
		SessionId: s.SessionID,
		Radius:    s.Radius,
		Profile:   s.Profile,
		Uptime:    s.Uptime,
	}
}

// toProtoPPPActiveSessions converts a slice of domain port.PPPActiveSession to protobuf PPPActiveSessions.
func toProtoPPPActiveSessions(sessions []port.PPPActiveSession) []*devicepb.PPPActiveSession {
	res := make([]*devicepb.PPPActiveSession, len(sessions))
	for i, s := range sessions {
		res[i] = toProtoPPPActiveSession(s)
	}
	return res
}

// FromProtoCreateSecretRequest maps CreatePPPSecretRequest to domain port.PPPoESecretParams.
func FromProtoCreateSecretRequest(req *devicepb.CreatePPPSecretRequest) port.PPPoESecretParams {
	return port.PPPoESecretParams{
		Name:          req.Name,
		Password:      req.Password,
		Profile:       req.Profile,
		Service:       req.Service,
		LocalAddress:  req.LocalAddress,
		RemoteAddress: req.RemoteAddress,
		Comment:       req.Comment,
		Disabled:      req.Disabled,
		CallerID:      req.CallerId,
	}
}

// FromProtoUpdateSecretRequest maps UpdatePPPSecretRequest to domain port.PPPoESecretParams.
func FromProtoUpdateSecretRequest(req *devicepb.UpdatePPPSecretRequest) port.PPPoESecretParams {
	return port.PPPoESecretParams{
		Name:          req.Name,
		Password:      req.Password,
		Profile:       req.Profile,
		Service:       req.Service,
		LocalAddress:  req.LocalAddress,
		RemoteAddress: req.RemoteAddress,
		Comment:       req.Comment,
		CallerID:      req.CallerId,
	}
}

// FromProtoCreateProfileRequest maps CreatePPPProfileRequest to domain port.PPPProfileParams.
func FromProtoCreateProfileRequest(req *devicepb.CreatePPPProfileRequest) port.PPPProfileParams {
	return port.PPPProfileParams{
		Name:           req.Name,
		RateLimit:      req.RateLimit,
		LocalAddress:   req.LocalAddress,
		RemoteAddress:  req.RemoteAddress,
		DNSServer:      req.DnsServer,
		ParentQueue:    req.ParentQueue,
		AddressList:    req.AddressList,
		Comment:        req.Comment,
		SharedUsers:    req.SharedUsers,
		OnlyOne:        req.OnlyOne,
		UseMPLS:        req.UseMpls,
		UseCompression: req.UseCompression,
		UseEncryption:  req.UseEncryption,
		ChangeTCPMSS:   req.ChangeTcpMss,
		BridgeLearning: req.BridgeLearning,
	}
}

// FromProtoUpdateProfileRequest maps UpdatePPPProfileRequest to domain port.PPPProfileParams.
func FromProtoUpdateProfileRequest(req *devicepb.UpdatePPPProfileRequest) port.PPPProfileParams {
	return port.PPPProfileParams{
		Name:           req.Name,
		RateLimit:      req.RateLimit,
		LocalAddress:   req.LocalAddress,
		RemoteAddress:  req.RemoteAddress,
		DNSServer:      req.DnsServer,
		ParentQueue:    req.ParentQueue,
		AddressList:    req.AddressList,
		Comment:        req.Comment,
		SharedUsers:    req.SharedUsers,
		OnlyOne:        req.OnlyOne,
		UseMPLS:        req.UseMpls,
		UseCompression: req.UseCompression,
		UseEncryption:  req.UseEncryption,
		ChangeTCPMSS:   req.ChangeTcpMss,
		BridgeLearning: req.BridgeLearning,
	}
}
