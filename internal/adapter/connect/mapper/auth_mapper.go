package mapper

import (
	"fmt"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/domain/auth"
)

// UserToProto maps an auth.User domain entity to a UserProfile Protobuf message.
func UserToProto(u *auth.User) *devicepb.UserProfile {
	if u == nil {
		return nil
	}
	return &devicepb.UserProfile{
		Id:       fmt.Sprintf("%d", u.ID),
		Username: u.Email,
		Email:    u.Email,
		Role:     u.Role,
	}
}

// LoginResultToProto maps an auth.LoginResult to a LoginResponse Protobuf message.
func LoginResultToProto(res *auth.LoginResult) *devicepb.LoginResponse {
	if res == nil {
		return nil
	}
	return &devicepb.LoginResponse{
		Token:         res.Token,
		User:          UserToProto(res.User),
		ExpiresAtUnix: res.ExpiresAtUnix,
	}
}
