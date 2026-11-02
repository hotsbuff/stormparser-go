package protocol

import (
	"bytes"
	"fmt"

	"github.com/hotsbuff/stormparser-go/internal/formats"
	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
)

func DecodeVersioned(data []byte, schema *Schema, typeID uint16) (*Value, error) {
	stream := kaitai.NewStream(bytes.NewReader(data))
	return decodeVersioned(stream, schema, typeID)
}

func decodeVersioned(stream *kaitai.Stream, schema *Schema, typeID uint16) (*Value, error) {
	raw := formats.NewStormValue()
	if err := raw.Read(stream, nil, raw); err != nil {
		return nil, fmt.Errorf("decode versioned value: %w", err)
	}
	value, err := rawValue(raw)
	if err != nil {
		return nil, err
	}
	return schema.bind(value, typeID)
}

func rawValue(raw *formats.StormValue) (*Value, error) {
	value := &Value{Kind: ValueKind(raw.Kind)}
	switch data := raw.Data.(type) {
	case *formats.StormValue_ArrayValue:
		value.Items = make([]*Value, len(data.Items))
		for i, item := range data.Items {
			converted, err := rawValue(item)
			if err != nil {
				return nil, err
			}
			value.Items[i] = converted
		}
	case *formats.StormValue_BitArrayValue:
		length, err := data.BitLength.Value()
		if err != nil {
			return nil, err
		}
		value.BitLength = int64(length)
		value.Bytes = data.Data
	case *formats.StormValue_BlobValue:
		value.Bytes = data.Data
	case *formats.StormValue_ChoiceValue:
		tag, err := data.Tag.Value()
		if err != nil {
			return nil, err
		}
		value.Tag = int64(tag)
		value.Value, err = rawValue(data.Value)
		if err != nil {
			return nil, err
		}
	case *formats.StormValue_OptionalValue:
		if data.Present != 0 {
			converted, err := rawValue(data.Value)
			if err != nil {
				return nil, err
			}
			value.Value = converted
		}
	case *formats.StormValue_StructValue:
		value.Fields = make([]*Value, len(data.Fields))
		for i, field := range data.Fields {
			tag, err := field.Tag.Value()
			if err != nil {
				return nil, err
			}
			converted, err := rawValue(field.Value)
			if err != nil {
				return nil, err
			}
			converted.Tag = int64(tag)
			value.Fields[i] = converted
		}
	case *formats.StormValue_BoolValue:
		value.Bool = data.Value != 0
	case *formats.StormValue_Bytes4Value:
		value.Bytes = data.Data
	case *formats.StormValue_Bytes8Value:
		value.Bytes = data.Data
	case *formats.StormValue_Vint:
		integer, err := data.Value()
		if err != nil {
			return nil, err
		}
		value.Int = int64(integer)
	default:
		return nil, fmt.Errorf("unknown versioned value kind %d", raw.Kind)
	}
	return value, nil
}

func (s *Schema) bind(raw *Value, typeID uint16) (*Value, error) {
	info, err := s.Type(typeID)
	if err != nil {
		return nil, err
	}
	result := *raw
	result.TypeName = string(info.Kind)

	switch info.Kind {
	case KindArray:
		result.Items = make([]*Value, len(raw.Items))
		for i, item := range raw.Items {
			result.Items[i], err = s.bind(item, info.TypeID)
			if err != nil {
				return nil, err
			}
		}
	case KindOptional:
		if raw.Value != nil {
			result.Value, err = s.bind(raw.Value, info.TypeID)
			if err != nil {
				return nil, err
			}
		}
	case KindStruct:
		fields := make(map[int64]Field, len(info.Fields))
		for _, field := range info.Fields {
			fields[field.Tag] = field
		}
		result.Fields = make([]*Value, len(raw.Fields))
		for i, field := range raw.Fields {
			result.Fields[i] = field
			if definition, ok := fields[field.Tag]; ok {
				result.Fields[i], err = s.bind(field, definition.TypeID)
				if err != nil {
					return nil, err
				}
				result.Fields[i].Name = definition.Name
				result.Fields[i].Tag = field.Tag
			}
		}
	case KindChoice:
		if definition, ok := info.Choices[raw.Tag]; ok {
			result.Name = definition.Name
			result.Value, err = s.bind(raw.Value, definition.TypeID)
			if err != nil {
				return nil, err
			}
		}
	case KindFourCC:
		result.Text = FourCC(raw.Bytes)
	case KindReal32:
		result.Float = Real32(raw.Bytes)
	case KindReal64:
		result.Float = Real64(raw.Bytes)
	}
	return &result, nil
}

func (v *Value) Field(name string) *Value {
	if v == nil || v.Kind != ValueStruct {
		return nil
	}
	for _, field := range v.Fields {
		if field.Name == name {
			return field
		}
		if field.Name == "__parent" {
			if nested := field.Field(name); nested != nil {
				return nested
			}
		}
	}
	return nil
}

func (v *Value) Unwrap() *Value {
	if v != nil && (v.Kind == ValueOptional || v.Kind == ValueChoice) {
		return v.Value.Unwrap()
	}
	return v
}
