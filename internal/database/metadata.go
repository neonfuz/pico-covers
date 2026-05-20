package database

type ConsoleType int

const (
	Unknown ConsoleType = iota
	NES
	SNES
	GB
	GBC
	GBA
	NDS
	DSi
	GameGear
	Genesis
	MasterSystem
	FDS
	NintendoDSDownloadPlay
	NintendoDSiDigital
)

var noIntroDBMapping = map[ConsoleType]ConsoleType{
	NintendoDSDownloadPlay: NDS,
	NintendoDSiDigital:     DSi,
}

func (c ConsoleType) BaseType() ConsoleType {
	if base, ok := noIntroDBMapping[c]; ok {
		return base
	}
	return c
}

type RomMetaData struct {
	ConsoleType    ConsoleType `json:"console_type"`
	ConsoleSubType ConsoleType `json:"console_sub_type"`
	Name           string      `json:"name"`
	Sha1           string      `json:"sha1"`
	Serial         string      `json:"serial"`
	GameId         string      `json:"game_id"`
	Status         string      `json:"status"`
}
