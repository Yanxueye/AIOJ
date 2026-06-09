package judger

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// codecName is the gRPC content-subtype advertised by both client and
// server. Registering our own codec lets us avoid generating protobuf
// bindings while still speaking real HTTP/2 + gRPC framing.
const codecName = "json"

// jsonCodec implements encoding.Codec using encoding/json. The payloads are
// plain Go structs tagged with `json:"..."` fields matching proto/judger.proto.
type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }

func (jsonCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }

func (jsonCodec) Name() string { return codecName }

func init() {
	encoding.RegisterCodec(jsonCodec{})
}
