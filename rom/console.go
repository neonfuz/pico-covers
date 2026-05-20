package rom

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

var consoleTypeStrings = map[ConsoleType]string{
	Unknown:                "Unknown",
	NES:                    "Nintendo Entertainment System",
	SNES:                   "Super Nintendo Entertainment System",
	GB:                     "Game Boy",
	GBC:                    "Game Boy Color",
	GBA:                    "Game Boy Advance",
	NDS:                    "Nintendo DS",
	DSi:                    "Nintendo DSi",
	GameGear:               "Sega Game Gear",
	Genesis:                "Sega Genesis",
	MasterSystem:           "Sega Master System",
	FDS:                    "Famicom Disk System",
	NintendoDSDownloadPlay: "Nintendo DS (Download Play)",
	NintendoDSiDigital:     "Nintendo DSi (Digital)",
}

func (c ConsoleType) String() string {
	if s, ok := consoleTypeStrings[c]; ok {
		return s
	}
	return "Unknown"
}

var NoIntroDBMapping = map[ConsoleType]ConsoleType{
	NintendoDSDownloadPlay: NDS,
	NintendoDSiDigital:     DSi,
}

func (c ConsoleType) BaseType() ConsoleType {
	if base, ok := NoIntroDBMapping[c]; ok {
		return base
	}
	return c
}

var libRetroConsoleStrings = map[ConsoleType]string{
	NES:          "Nintendo_-_Nintendo_Entertainment_System",
	SNES:         "Nintendo_-_Super_Nintendo_Entertainment_System",
	GB:           "Nintendo_-_Game_Boy",
	GBC:          "Nintendo_-_Game_Boy_Color",
	GBA:          "Nintendo_-_Game_Boy_Advance",
	NDS:          "Nintendo_-_Nintendo_DS",
	DSi:          "Nintendo_-_Nintendo_DSi",
	GameGear:     "Sega_-_Game_Gear",
	Genesis:      "Sega_-_Mega_Drive_-_Genesis",
	MasterSystem: "Sega_-_Master_System_-_Mark_III",
	FDS:          "Nintendo_-_Family_Computer_Disk_System",
}

func (c ConsoleType) LibRetroConsoleString() string {
	if s, ok := libRetroConsoleStrings[c.BaseType()]; ok {
		return s
	}
	return ""
}

var ExtensionMapping = map[string]ConsoleType{
	".nes":  NES,
	".sfc":  SNES,
	".smc":  SNES,
	".snes": SNES,
	".gb":   GB,
	".sgb":  GB,
	".gbc":  GBC,
	".gba":  GBA,
	".nds":  NDS,
	".ds":   NDS,
	".dsi":  DSi,
	".gg":   GameGear,
	".gen":  Genesis,
	".sms":  MasterSystem,
	".fds":  FDS,
}

func ConsoleTypeFromExtension(ext string) (ConsoleType, bool) {
	ct, ok := ExtensionMapping[ext]
	return ct, ok
}
