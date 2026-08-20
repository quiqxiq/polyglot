package connect

import (
	"encoding/json"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type protoJSONCodec struct{}

func (protoJSONCodec) Name() string { return "json" }

func (protoJSONCodec) Marshal(v any) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		return protojson.MarshalOptions{
			UseProtoNames:   false,
			EmitUnpopulated: true,
		}.Marshal(m)
	}
	return json.Marshal(v)
}

func (protoJSONCodec) Unmarshal(b []byte, v any) error {
	if m, ok := v.(proto.Message); ok {
		return protojson.UnmarshalOptions{
			DiscardUnknown: true,
		}.Unmarshal(b, m)
	}
	return json.Unmarshal(b, v)
}

// JSONCodec returns a ConnectRPC codec that uses protojson for Protobuf messages.
func JSONCodec() connect.Codec {
	return protoJSONCodec{}
}
