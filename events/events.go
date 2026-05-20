package events

type EventKind string

const (
	EventROMStart    EventKind = "rom_start"
	EventROMSuccess  EventKind = "rom_success"
	EventROMSkipped  EventKind = "rom_skipped"
	EventROMNotFound EventKind = "rom_not_found"
	EventROMError    EventKind = "rom_error"
	EventDBInit      EventKind = "db_init"
	EventDBLoaded    EventKind = "db_loaded"
	EventInfo        EventKind = "info"
)

func (k EventKind) String() string {
	return string(k)
}

type ProgressEvent struct {
	Kind      EventKind
	ROMFile   string
	GameTitle string
	Detail    string
	Total     int
	Completed int
}

type EventHandler func(ProgressEvent)
