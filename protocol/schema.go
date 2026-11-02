package protocol

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Kind string

const (
	KindArray    Kind = "array"
	KindBitArray Kind = "bitarray"
	KindBlob     Kind = "blob"
	KindBool     Kind = "bool"
	KindChoice   Kind = "choice"
	KindFourCC   Kind = "fourcc"
	KindInt      Kind = "int"
	KindNull     Kind = "null"
	KindOptional Kind = "optional"
	KindReal32   Kind = "real32"
	KindReal64   Kind = "real64"
	KindStruct   Kind = "struct"
)

type Bounds struct {
	Offset int64 `yaml:"offset" json:"offset"`
	Bits   uint8 `yaml:"bits" json:"bits"`
}

type Field struct {
	Name   string `yaml:"name" json:"name"`
	TypeID uint16 `yaml:"type" json:"type"`
	Tag    int64  `yaml:"tag" json:"tag"`
}

type TypeInfo struct {
	Kind    Kind            `yaml:"kind" json:"kind"`
	Bounds  Bounds          `yaml:"bounds" json:"bounds"`
	TypeID  uint16          `yaml:"type,omitempty" json:"type,omitempty"`
	Fields  []Field         `yaml:"fields,omitempty" json:"fields,omitempty"`
	Choices map[int64]Field `yaml:"choices,omitempty" json:"choices,omitempty"`
}

type EventType struct {
	TypeID uint16 `yaml:"type" json:"type"`
	Name   string `yaml:"name" json:"name"`
}

type Schema struct {
	Build              uint32              `yaml:"build" json:"build"`
	Extends            uint32              `yaml:"extends,omitempty" json:"extends,omitempty"`
	Exact              bool                `yaml:"exact" json:"exact"`
	Types              []TypeInfo          `yaml:"types" json:"types"`
	HeaderType         uint16              `yaml:"header_type" json:"header_type"`
	DetailsType        uint16              `yaml:"details_type" json:"details_type"`
	InitDataType       uint16              `yaml:"init_data_type" json:"init_data_type"`
	GameEventIDType    uint16              `yaml:"game_event_id_type" json:"game_event_id_type"`
	MessageEventIDType uint16              `yaml:"message_event_id_type" json:"message_event_id_type"`
	TrackerEventIDType uint16              `yaml:"tracker_event_id_type" json:"tracker_event_id_type"`
	DeltaType          uint16              `yaml:"delta_type" json:"delta_type"`
	UserIDType         uint16              `yaml:"user_id_type" json:"user_id_type"`
	GameEvents         map[int64]EventType `yaml:"game_events,omitempty" json:"game_events,omitempty"`
	MessageEvents      map[int64]EventType `yaml:"message_events,omitempty" json:"message_events,omitempty"`
	TrackerEvents      map[int64]EventType `yaml:"tracker_events,omitempty" json:"tracker_events,omitempty"`
}

func ParseSchema(data []byte) (*Schema, error) {
	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse protocol schema: %w", err)
	}
	return &schema, nil
}

func (s *Schema) Type(id uint16) (TypeInfo, error) {
	if int(id) >= len(s.Types) {
		return TypeInfo{}, fmt.Errorf("protocol build %d: unknown type %d", s.Build, id)
	}
	return s.Types[id], nil
}
