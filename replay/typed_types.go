package replay

import (
	"time"

	"github.com/hotsbuff/stormparser-go/protocol"
)

type HeaderInfo struct {
	Signature        string          `json:"signature"`
	Version          Version         `json:"version"`
	Type             int64           `json:"type"`
	ElapsedGameLoops int64           `json:"elapsed_game_loops"`
	Duration         time.Duration   `json:"duration"`
	UseScaledTime    bool            `json:"use_scaled_time"`
	DataBuildNumber  int64           `json:"data_build_number"`
	Raw              *protocol.Value `json:"-"`
}

type Details struct {
	Players         []Player        `json:"players"`
	Title           string          `json:"title"`
	Difficulty      string          `json:"difficulty"`
	MapFileName     string          `json:"map_file_name"`
	IsBlizzardMap   bool            `json:"is_blizzard_map"`
	TimeUTC         time.Time       `json:"time_utc"`
	TimeLocal       time.Time       `json:"time_local"`
	TimeLocalOffset time.Duration   `json:"time_local_offset"`
	GameSpeed       int64           `json:"game_speed"`
	Raw             *protocol.Value `json:"-"`
}

type Player struct {
	Name             string          `json:"name"`
	Toon             Toon            `json:"toon"`
	Race             string          `json:"race"`
	Color            Color           `json:"color"`
	Control          int64           `json:"control"`
	TeamID           int64           `json:"team_id"`
	Handicap         int64           `json:"handicap"`
	Observe          int64           `json:"observe"`
	Result           int64           `json:"result"`
	WorkingSetSlotID int64           `json:"working_set_slot_id"`
	Hero             string          `json:"hero"`
	Raw              *protocol.Value `json:"-"`
}

type Toon struct {
	Region    int64  `json:"region"`
	ProgramID string `json:"program_id"`
	Realm     int64  `json:"realm"`
	Name      string `json:"name"`
	ID        int64  `json:"id"`
}

type Color struct {
	A int64 `json:"a"`
	R int64 `json:"r"`
	G int64 `json:"g"`
	B int64 `json:"b"`
}

type Summary struct {
	Version       Version       `json:"version"`
	ProtocolBuild uint32        `json:"protocol_build"`
	ExactProtocol bool          `json:"exact_protocol"`
	Map           string        `json:"map"`
	PlayedAt      time.Time     `json:"played_at"`
	Duration      time.Duration `json:"duration"`
	Players       []Player      `json:"players"`
}
