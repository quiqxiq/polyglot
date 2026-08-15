package mapper

import (
	devicepb "github.com/quixiq/polyglot/api/proto/v1"
)

// PolicyRowsToProto maps raw string slices of Casbin rules to Protobuf Policy messages.
func PolicyRowsToProto(rawPolicies [][]string) []*devicepb.Policy {
	policies := make([]*devicepb.Policy, 0, len(rawPolicies))
	for _, p := range rawPolicies {
		if len(p) >= 3 {
			policies = append(policies, &devicepb.Policy{
				Sub: p[0],
				Obj: p[1],
				Act: p[2],
			})
		}
	}
	return policies
}
