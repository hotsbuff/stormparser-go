package protocol

import (
	"bytes"
	"fmt"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
)

type Event struct {
	Name     string `json:"name"`
	ID       int64  `json:"id"`
	GameLoop int64  `json:"gameloop"`
	UserID   *int64 `json:"userid,omitempty"`
	Payload  *Value `json:"payload"`
}

func DecodeGameEvents(data []byte, schema *Schema) ([]Event, error) {
	return decodeBitPackedEvents(data, schema, schema.GameEventIDType, schema.GameEvents, true)
}

func DecodeMessageEvents(data []byte, schema *Schema) ([]Event, error) {
	return decodeBitPackedEvents(data, schema, schema.MessageEventIDType, schema.MessageEvents, true)
}

func decodeBitPackedEvents(data []byte, schema *Schema, eventIDType uint16, eventTypes map[int64]EventType, withUser bool) ([]Event, error) {
	decoder := NewBitPackedDecoder(data, schema)
	var gameLoop int64
	var events []Event
	for !decoder.Buffer.Done() {
		delta, err := decoder.Decode(schema.DeltaType)
		if err != nil {
			return nil, fmt.Errorf("decode event gameloop: %w", err)
		}
		gameLoop += delta.Unwrap().Int
		var userID *int64
		if withUser {
			user, err := decoder.Decode(schema.UserIDType)
			if err != nil {
				return nil, fmt.Errorf("decode event userid: %w", err)
			}
			id := user.Field("m_userId").Int
			userID = &id
		}
		eventID, err := decoder.Decode(eventIDType)
		if err != nil {
			return nil, fmt.Errorf("decode event id: %w", err)
		}
		eventType, ok := eventTypes[eventID.Int]
		if !ok {
			return nil, fmt.Errorf("unknown event id %d at gameloop %d", eventID.Int, gameLoop)
		}
		payload, err := decoder.Decode(eventType.TypeID)
		if err != nil {
			return nil, fmt.Errorf("decode %s at gameloop %d: %w", eventType.Name, gameLoop, err)
		}
		events = append(events, Event{Name: eventType.Name, ID: eventID.Int, GameLoop: gameLoop, UserID: userID, Payload: payload})
		decoder.Buffer.ByteAlign()
	}
	return events, nil
}

func DecodeTrackerEvents(data []byte, schema *Schema) ([]Event, error) {
	stream := kaitai.NewStream(bytes.NewReader(data))
	var gameLoop int64
	var events []Event
	for {
		position, err := stream.Pos()
		if err != nil {
			return nil, err
		}
		if position >= int64(len(data)) {
			break
		}
		delta, err := decodeVersioned(stream, schema, schema.DeltaType)
		if err != nil {
			return nil, fmt.Errorf("decode tracker gameloop: %w", err)
		}
		gameLoop += delta.Unwrap().Int
		eventID, err := decodeVersioned(stream, schema, schema.TrackerEventIDType)
		if err != nil {
			return nil, fmt.Errorf("decode tracker event id: %w", err)
		}
		eventType, ok := schema.TrackerEvents[eventID.Int]
		if !ok {
			return nil, fmt.Errorf("unknown tracker event id %d at gameloop %d", eventID.Int, gameLoop)
		}
		payload, err := decodeVersioned(stream, schema, eventType.TypeID)
		if err != nil {
			return nil, fmt.Errorf("decode %s at gameloop %d: %w", eventType.Name, gameLoop, err)
		}
		events = append(events, Event{Name: eventType.Name, ID: eventID.Int, GameLoop: gameLoop, Payload: payload})
	}
	return events, nil
}
