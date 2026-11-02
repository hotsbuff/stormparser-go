package protocol

import (
	"bytes"
	"fmt"

	"github.com/hotsbuff/stormparser-go/internal/formats"
	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
)

type Attributes struct {
	Source       uint8       `json:"source"`
	MapNamespace uint32      `json:"map_namespace"`
	Entries      []Attribute `json:"entries"`
}

type Attribute struct {
	Namespace uint32 `json:"namespace"`
	ID        uint32 `json:"id"`
	Scope     uint8  `json:"scope"`
	Value     string `json:"value"`
}

func DecodeAttributes(data []byte) (*Attributes, error) {
	stream := kaitai.NewStream(bytes.NewReader(data))
	decoded := formats.NewAttributesEvents()
	if err := decoded.Read(stream, nil, decoded); err != nil {
		return nil, fmt.Errorf("decode attributes: %w", err)
	}
	result := &Attributes{Source: decoded.Source, MapNamespace: decoded.MapNamespace}
	for _, entry := range decoded.Entries {
		value := make([]byte, 4)
		for i := range entry.Value {
			value[len(entry.Value)-1-i] = entry.Value[i]
		}
		value = bytes.Trim(value, "\x00")
		result.Entries = append(result.Entries, Attribute{
			Namespace: entry.Namespace,
			ID:        entry.Id,
			Scope:     entry.Scope,
			Value:     string(value),
		})
	}
	return result, nil
}
