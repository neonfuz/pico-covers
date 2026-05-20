package database

import (
	"testing"
)

func TestFindBySHA1(t *testing.T) {
	db := NewDatabase()
	db.Records = []RomMetaData{
		{Name: "Game A", Sha1: "abc123", ConsoleType: NDS, ConsoleSubType: NDS},
		{Name: "Game B", Sha1: "def456", ConsoleType: GBA, ConsoleSubType: GBA},
		{Name: "Game C", Sha1: "abc123", ConsoleType: NDS, ConsoleSubType: NDS},
	}
	db.buildIndex()

	result := db.FindBySHA1("abc123")
	if result == nil {
		t.Fatal("expected match for abc123")
	}
	if result.Name != "Game A" {
		t.Errorf("expected 'Game A', got %q", result.Name)
	}

	result = db.FindBySHA1("DEF456")
	if result == nil {
		t.Fatal("expected case-insensitive match for DEF456")
	}
	if result.Name != "Game B" {
		t.Errorf("expected 'Game B', got %q", result.Name)
	}

	result = db.FindBySHA1("nonexistent")
	if result != nil {
		t.Errorf("expected nil for nonexistent SHA1, got %v", result)
	}
}

func TestFindByTitleID(t *testing.T) {
	db := NewDatabase()
	db.Records = []RomMetaData{
		{Name: "Good Game", Serial: "ABCD", GameId: "XYZ", ConsoleType: NDS, ConsoleSubType: NDS, Status: "verified"},
		{Name: "Bad Game", Serial: "ABCD", GameId: "", ConsoleType: NDS, ConsoleSubType: NDS, Status: "bad"},
		{Name: "Other Game", Serial: "EFGH", GameId: "", ConsoleType: GBA, ConsoleSubType: GBA, Status: "verified"},
	}
	db.buildIndex()

	result := db.FindByTitleID("ABCD", NDS)
	if result == nil {
		t.Fatal("expected match for ABCD on NDS")
	}
	if result.Name != "Good Game" {
		t.Errorf("expected 'Good Game' (non-bad), got %q", result.Name)
	}

	result = db.FindByTitleID("efgh", GBA)
	if result == nil {
		t.Fatal("expected case-insensitive match for efgh on GBA")
	}
	if result.Name != "Other Game" {
		t.Errorf("expected 'Other Game', got %q", result.Name)
	}

	result = db.FindByTitleID("ABCD", GBA)
	if result != nil {
		t.Errorf("expected nil for wrong console type match, got %v", result)
	}

	result = db.FindByTitleID("", NDS)
	if result != nil {
		t.Errorf("expected nil for empty title ID")
	}
}

func TestBaseTypeMapping(t *testing.T) {
	if NintendoDSDownloadPlay.BaseType() != NDS {
		t.Errorf("expected NintendoDSDownloadPlay -> NDS")
	}
	if NintendoDSiDigital.BaseType() != DSi {
		t.Errorf("expected NintendoDSiDigital -> DSi")
	}
	if NDS.BaseType() != NDS {
		t.Errorf("expected NDS -> NDS (no change)")
	}
}

func TestLibRetroParse(t *testing.T) {
	input := `game (
	name "Adventure Island (USA)"
	rom ( name "Adventure Island (USA).nes" size 262144 crc 1234ABCD sha1 abc123def456 )
)
game (
	name "Super Mario Bros. (World)"
	rom ( name "Super Mario Bros. (World).nes" size 40976 sha1 fed456cba789 )
)`

	results, err := parseSExpr(input, NES)
	if err != nil {
		t.Fatalf("parseSExpr failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Name != "Adventure Island (USA)" {
		t.Errorf("unexpected name: %q", results[0].Name)
	}
	if results[0].Sha1 != "abc123def456" {
		t.Errorf("unexpected sha1: %q", results[0].Sha1)
	}
	if results[0].ConsoleType != NES {
		t.Errorf("unexpected console type: %v", results[0].ConsoleType)
	}
	if results[1].Name != "Super Mario Bros. (World)" {
		t.Errorf("unexpected name: %q", results[1].Name)
	}
	if results[1].Sha1 != "fed456cba789" {
		t.Errorf("unexpected sha1: %q", results[1].Sha1)
	}
}
