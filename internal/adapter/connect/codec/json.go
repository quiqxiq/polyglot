package codec

import (
	"encoding/json"

	"connectrpc.com/connect"
)

// JSONCodec implements the connect.Codec interface for JSON serialization.
type JSONCodec struct{}

func (JSONCodec) Name() string                     { return "json" }
func (JSONCodec) Marshal(v any) ([]byte, error)    { return json.Marshal(v) }
func (JSONCodec) Unmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

// Option returns the Connect handler option configured with JSONCodec.
func Option() connect.HandlerOption {
	return connect.WithCodec(JSONCodec{})
}
