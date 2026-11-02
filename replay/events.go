package replay

import "github.com/hotsbuff/stormparser-go/protocol"

func (r *Replay) GameEvents() ([]protocol.Event, error) {
	data, err := r.ReadFile(GameEventsFile)
	if err != nil {
		return nil, err
	}
	return protocol.DecodeGameEvents(data, r.Schema)
}

func (r *Replay) MessageEvents() ([]protocol.Event, error) {
	data, err := r.ReadFile(MessageEventsFile)
	if err != nil {
		return nil, err
	}
	return protocol.DecodeMessageEvents(data, r.Schema)
}

func (r *Replay) TrackerEvents() ([]protocol.Event, error) {
	data, err := r.ReadFile(TrackerEventsFile)
	if err != nil {
		return nil, err
	}
	return protocol.DecodeTrackerEvents(data, r.Schema)
}
