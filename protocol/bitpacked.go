package protocol

import (
	"encoding/binary"
	"fmt"
	"math"
)

type BitPackedBuffer struct {
	data     []byte
	used     int
	next     uint8
	nextBits uint8
}

func NewBitPackedBuffer(data []byte) *BitPackedBuffer {
	return &BitPackedBuffer{data: data}
}

func (b *BitPackedBuffer) Done() bool {
	return b.nextBits == 0 && b.used >= len(b.data)
}

func (b *BitPackedBuffer) ByteAlign() {
	b.nextBits = 0
}

func (b *BitPackedBuffer) ReadBits(count uint) (uint64, error) {
	if count > 64 {
		return 0, fmt.Errorf("cannot read %d bits into uint64", count)
	}
	var result uint64
	resultBits := uint(0)
	for resultBits < count {
		if b.nextBits == 0 {
			if b.used >= len(b.data) {
				return 0, fmt.Errorf("bit-packed buffer is truncated")
			}
			b.next = b.data[b.used]
			b.used++
			b.nextBits = 8
		}
		copyBits := uint(b.nextBits)
		if remaining := count - resultBits; remaining < copyBits {
			copyBits = remaining
		}
		mask := uint16(1<<copyBits) - 1
		copy := uint64(b.next) & uint64(mask)
		result |= copy << (count - resultBits - copyBits)
		b.next >>= copyBits
		b.nextBits -= uint8(copyBits)
		resultBits += copyBits
	}
	return result, nil
}

func (b *BitPackedBuffer) ReadBitArray(count uint) ([]byte, error) {
	result := make([]byte, (count+7)/8)
	for bitIndex := uint(0); bitIndex < count; bitIndex++ {
		bit, err := b.ReadBits(1)
		if err != nil {
			return nil, err
		}
		if bit != 0 {
			result[bitIndex/8] |= 1 << (7 - bitIndex%8)
		}
	}
	return result, nil
}

func (b *BitPackedBuffer) ReadAligned(count int) ([]byte, error) {
	b.ByteAlign()
	if count < 0 || count > len(b.data)-b.used {
		return nil, fmt.Errorf("bit-packed buffer is truncated: need %d bytes", count)
	}
	result := b.data[b.used : b.used+count]
	b.used += count
	return result, nil
}

func (b *BitPackedBuffer) ReadUnaligned(count int) ([]byte, error) {
	result := make([]byte, count)
	for i := range result {
		value, err := b.ReadBits(8)
		if err != nil {
			return nil, err
		}
		result[i] = byte(value)
	}
	return result, nil
}

type BitPackedDecoder struct {
	Buffer *BitPackedBuffer
	Schema *Schema
}

func NewBitPackedDecoder(data []byte, schema *Schema) *BitPackedDecoder {
	return &BitPackedDecoder{Buffer: NewBitPackedBuffer(data), Schema: schema}
}

func DecodeBitPacked(data []byte, schema *Schema, typeID uint16) (*Value, error) {
	decoder := NewBitPackedDecoder(data, schema)
	value, err := decoder.Decode(typeID)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (d *BitPackedDecoder) Decode(typeID uint16) (*Value, error) {
	info, err := d.Schema.Type(typeID)
	if err != nil {
		return nil, err
	}
	value := &Value{TypeName: string(info.Kind)}
	switch info.Kind {
	case KindArray:
		length, err := d.readInt(info.Bounds)
		if err != nil {
			return nil, err
		}
		if length < 0 || length > int64(d.Buffer.used+len(d.Buffer.data))+1 {
			return nil, fmt.Errorf("invalid bit-packed array length %d", length)
		}
		value.Kind = ValueArray
		value.Items = make([]*Value, length)
		for i := range value.Items {
			value.Items[i], err = d.Decode(info.TypeID)
			if err != nil {
				return nil, err
			}
		}
	case KindBitArray:
		length, err := d.readInt(info.Bounds)
		if err != nil {
			return nil, err
		}
		bits, err := d.Buffer.ReadBitArray(uint(length))
		if err != nil {
			return nil, err
		}
		value.Kind = ValueBitArray
		value.BitLength = length
		value.Bytes = bits
		if length <= 64 {
			for _, b := range bits {
				value.Int = value.Int<<8 | int64(b)
			}
		}
	case KindBlob:
		length, err := d.readInt(info.Bounds)
		if err != nil {
			return nil, err
		}
		value.Kind = ValueBlob
		value.Bytes, err = d.Buffer.ReadAligned(int(length))
		if err != nil {
			return nil, err
		}
	case KindBool:
		integer, err := d.readInt(Bounds{Bits: 1})
		if err != nil {
			return nil, err
		}
		value.Kind = ValueBool
		value.Bool = integer != 0
	case KindChoice:
		tag, err := d.readInt(info.Bounds)
		if err != nil {
			return nil, err
		}
		field, ok := info.Choices[tag]
		if !ok {
			return nil, fmt.Errorf("unknown bit-packed choice tag %d for type %d", tag, typeID)
		}
		value.Kind = ValueChoice
		value.Tag = tag
		value.Name = field.Name
		value.Value, err = d.Decode(field.TypeID)
		if err != nil {
			return nil, err
		}
	case KindFourCC:
		bits, err := d.Buffer.ReadBits(32)
		if err != nil {
			return nil, err
		}
		value.Kind = ValueBytes4
		value.Bytes = []byte{byte(bits >> 24), byte(bits >> 16), byte(bits >> 8), byte(bits)}
		value.Text = FourCC(value.Bytes)
	case KindInt:
		integer, err := d.readInt(info.Bounds)
		if err != nil {
			return nil, err
		}
		value.Kind = ValueInt
		value.Int = integer
	case KindNull:
		value.Kind = ValueNull
	case KindOptional:
		exists, err := d.readInt(Bounds{Bits: 1})
		if err != nil {
			return nil, err
		}
		value.Kind = ValueOptional
		if exists != 0 {
			value.Value, err = d.Decode(info.TypeID)
			if err != nil {
				return nil, err
			}
		}
	case KindReal32:
		data, err := d.Buffer.ReadUnaligned(4)
		if err != nil {
			return nil, err
		}
		value.Kind = ValueBytes4
		value.Float = float64(math.Float32frombits(binary.BigEndian.Uint32(data)))
	case KindReal64:
		data, err := d.Buffer.ReadUnaligned(8)
		if err != nil {
			return nil, err
		}
		value.Kind = ValueBytes8
		value.Float = math.Float64frombits(binary.BigEndian.Uint64(data))
	case KindStruct:
		value.Kind = ValueStruct
		value.Fields = make([]*Value, 0, len(info.Fields))
		for _, field := range info.Fields {
			fieldValue, err := d.Decode(field.TypeID)
			if err != nil {
				return nil, fmt.Errorf("decode field %s: %w", field.Name, err)
			}
			fieldValue.Name = field.Name
			fieldValue.Tag = field.Tag
			value.Fields = append(value.Fields, fieldValue)
		}
	default:
		return nil, fmt.Errorf("unsupported bit-packed type %q", info.Kind)
	}
	return value, nil
}

func (d *BitPackedDecoder) readInt(bounds Bounds) (int64, error) {
	bits, err := d.Buffer.ReadBits(uint(bounds.Bits))
	if err != nil {
		return 0, err
	}
	return bounds.Offset + int64(bits), nil
}
