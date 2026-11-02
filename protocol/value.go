package protocol

import "math"

type ValueKind uint8

const (
	ValueArray    ValueKind = 0
	ValueBitArray ValueKind = 1
	ValueBlob     ValueKind = 2
	ValueChoice   ValueKind = 3
	ValueOptional ValueKind = 4
	ValueStruct   ValueKind = 5
	ValueBool     ValueKind = 6
	ValueBytes4   ValueKind = 7
	ValueBytes8   ValueKind = 8
	ValueInt      ValueKind = 9
	ValueNull     ValueKind = 255
)

type Value struct {
	Kind      ValueKind `json:"kind"`
	TypeName  string    `json:"type_name,omitempty"`
	Name      string    `json:"name,omitempty"`
	Tag       int64     `json:"tag,omitempty"`
	Int       int64     `json:"int,omitempty"`
	Bool      bool      `json:"bool,omitempty"`
	Bytes     []byte    `json:"bytes,omitempty"`
	Text      string    `json:"text,omitempty"`
	BitLength int64     `json:"bit_length,omitempty"`
	Float     float64   `json:"float,omitempty"`
	Items     []*Value  `json:"items,omitempty"`
	Fields    []*Value  `json:"fields,omitempty"`
	Value     *Value    `json:"value,omitempty"`
}

type BitArray struct {
	BitLength int64
	Bytes     []byte
}

func FourCC(data []byte) string {
	return string(data)
}

func Real32(data []byte) float64 {
	return float64(math.Float32frombits(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])))
}

func Real64(data []byte) float64 {
	bits := uint64(0)
	for _, b := range data {
		bits = bits<<8 | uint64(b)
	}
	return math.Float64frombits(bits)
}
