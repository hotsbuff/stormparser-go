package replay

import (
	"fmt"
	"time"

	"github.com/hotsbuff/stormparser-go/protocol"
)

const windowsToUnixEpochTicks = 116444736000000000

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Revision, v.Build)
}

func (r *Replay) HeaderInfo() HeaderInfo {
	loops := intField(r.Header, "m_elapsedGameLoops")
	return HeaderInfo{
		Signature:        stringField(r.Header, "m_signature"),
		Version:          r.Version(),
		Type:             intField(r.Header, "m_type"),
		ElapsedGameLoops: loops,
		Duration:         time.Duration(loops) * time.Second / 16,
		UseScaledTime:    boolField(r.Header, "m_useScaledTime"),
		DataBuildNumber:  intField(r.Header, "m_dataBuildNum"),
		Raw:              r.Header,
	}
}

func (r *Replay) TypedDetails() (*Details, error) {
	raw, err := r.Details()
	if err != nil {
		return nil, err
	}
	details := &Details{
		Title:         stringField(raw, "m_title"),
		Difficulty:    stringField(raw, "m_difficulty"),
		MapFileName:   stringField(raw, "m_mapFileName"),
		IsBlizzardMap: boolField(raw, "m_isBlizzardMap"),
		GameSpeed:     intField(raw, "m_gameSpeed"),
		Raw:           raw,
	}
	details.TimeUTC = windowsTime(intField(raw, "m_timeUTC"))
	details.TimeLocal = windowsTime(intField(raw, "m_timeLocalOffset"))
	if !details.TimeLocal.IsZero() && !details.TimeUTC.IsZero() {
		details.TimeLocalOffset = details.TimeLocal.Sub(details.TimeUTC)
	}
	players := raw.Field("m_playerList").Unwrap()
	if players != nil {
		details.Players = make([]Player, 0, len(players.Items))
		for _, rawPlayer := range players.Items {
			details.Players = append(details.Players, typedPlayer(rawPlayer))
		}
	}
	return details, nil
}

func (r *Replay) Summary() (*Summary, error) {
	details, err := r.TypedDetails()
	if err != nil {
		return nil, err
	}
	return &Summary{
		Version:       r.Version(),
		ProtocolBuild: r.Schema.Build,
		ExactProtocol: r.Schema.Exact,
		Map:           details.Title,
		PlayedAt:      details.TimeUTC,
		Duration:      r.HeaderInfo().Duration,
		Players:       details.Players,
	}, nil
}

func typedPlayer(raw *protocol.Value) Player {
	toon := raw.Field("m_toon")
	color := raw.Field("m_color")
	return Player{
		Name: stringField(raw, "m_name"),
		Toon: Toon{
			Region:    intField(toon, "m_region"),
			ProgramID: textField(toon, "m_programId"),
			Realm:     intField(toon, "m_realm"),
			Name:      stringField(toon, "m_name"),
			ID:        intField(toon, "m_id"),
		},
		Race: stringField(raw, "m_race"),
		Color: Color{
			A: intField(color, "m_a"),
			R: intField(color, "m_r"),
			G: intField(color, "m_g"),
			B: intField(color, "m_b"),
		},
		Control:          intField(raw, "m_control"),
		TeamID:           intField(raw, "m_teamId"),
		Handicap:         intField(raw, "m_handicap"),
		Observe:          intField(raw, "m_observe"),
		Result:           intField(raw, "m_result"),
		WorkingSetSlotID: intField(raw, "m_workingSetSlotId"),
		Hero:             stringField(raw, "m_hero"),
		Raw:              raw,
	}
}

func intField(value *protocol.Value, name string) int64 {
	field := value.Field(name).Unwrap()
	if field == nil {
		return 0
	}
	return field.Int
}

func boolField(value *protocol.Value, name string) bool {
	field := value.Field(name).Unwrap()
	return field != nil && field.Bool
}

func stringField(value *protocol.Value, name string) string {
	field := value.Field(name).Unwrap()
	if field == nil {
		return ""
	}
	if field.Text != "" {
		return field.Text
	}
	return string(field.Bytes)
}

func textField(value *protocol.Value, name string) string {
	return stringField(value, name)
}

func windowsTime(ticks int64) time.Time {
	if ticks <= windowsToUnixEpochTicks {
		return time.Time{}
	}
	return time.Unix((ticks-windowsToUnixEpochTicks)/10_000_000, 0).UTC()
}
