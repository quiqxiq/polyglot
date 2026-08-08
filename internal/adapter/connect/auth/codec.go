package auth

import "encoding/json"

type connectJSONCodec struct{}

func (connectJSONCodec) Name() string                     { return "json" }
func (connectJSONCodec) Marshal(v any) ([]byte, error)    { return json.Marshal(v) }
func (connectJSONCodec) Unmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }
