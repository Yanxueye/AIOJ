package grpcjson

import "encoding/json"

// Codec 提供 gRPC JSON 编解码。
type Codec struct{}

// Name 返回编解码器名称。
func (Codec) Name() string {
	return "json"
}

// Marshal 将对象编码为 JSON。
func (Codec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal 将 JSON 解码到对象。
func (Codec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
