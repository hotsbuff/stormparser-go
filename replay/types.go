package replay

const (
	HeaderFile        = "replay.header"
	DetailsFile       = "replay.details"
	InitDataFile      = "replay.initData"
	GameEventsFile    = "replay.game.events"
	MessageEventsFile = "replay.message.events"
	TrackerEventsFile = "replay.tracker.events"
	AttributesFile    = "replay.attributes.events"
)

type Version struct {
	Flags     int64 `json:"flags"`
	Major     int64 `json:"major"`
	Minor     int64 `json:"minor"`
	Revision  int64 `json:"revision"`
	Build     int64 `json:"build"`
	BaseBuild int64 `json:"base_build"`
}
