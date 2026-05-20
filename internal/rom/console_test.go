package rom

import "testing"

func TestGBHeaderDetection(t *testing.T) {
	header := make([]byte, 0x200)
	copy(header[0x104:], []byte{0xCE, 0xED, 0x66, 0x66})
	header[0x143] = 0x00

	rom, err := fromData(header, "test.gb")
	if err != nil {
		t.Fatalf("fromData failed: %v", err)
	}
	if rom.ConsoleType != GB {
		t.Errorf("expected GB, got %v", rom.ConsoleType)
	}
}

func TestGBCHeaderDetection(t *testing.T) {
	header := make([]byte, 0x200)
	copy(header[0x104:], []byte{0xCE, 0xED, 0x66, 0x66})
	header[0x143] = 0x80

	rom, err := fromData(header, "test.gbc")
	if err != nil {
		t.Fatalf("fromData failed: %v", err)
	}
	if rom.ConsoleType != GBC {
		t.Errorf("expected GBC, got %v", rom.ConsoleType)
	}
}

func TestGBCC0HeaderDetection(t *testing.T) {
	header := make([]byte, 0x200)
	copy(header[0x104:], []byte{0xCE, 0xED, 0x66, 0x66})
	header[0x143] = 0xC0

	rom, err := fromData(header, "test.gbc")
	if err != nil {
		t.Fatalf("fromData failed: %v", err)
	}
	if rom.ConsoleType != GBC {
		t.Errorf("expected GBC, got %v", rom.ConsoleType)
	}
}

func TestGBAHeaderDetection(t *testing.T) {
	header := make([]byte, 0x200)
	copy(header[0x4:], []byte{0x24, 0xFF, 0xAE, 0x51})
	titleId := "ABCE"
	copy(header[0xAC:], []byte(titleId))

	rom, err := fromData(header, "test.gba")
	if err != nil {
		t.Fatalf("fromData failed: %v", err)
	}
	if rom.ConsoleType != GBA {
		t.Errorf("expected GBA, got %v", rom.ConsoleType)
	}
	if rom.TitleId != titleId {
		t.Errorf("expected TitleId %q, got %q", titleId, rom.TitleId)
	}
}

func TestNDSHeaderDetection(t *testing.T) {
	header := make([]byte, 0x200)
	copy(header[0xC0:], []byte{0x24, 0xFF, 0xAE, 0x51})
	header[0x12] = 0x00
	titleId := "ABCD"
	copy(header[0x0C:], []byte(titleId))

	rom, err := fromData(header, "test.nds")
	if err != nil {
		t.Fatalf("fromData failed: %v", err)
	}
	if rom.ConsoleType != NDS {
		t.Errorf("expected NDS, got %v", rom.ConsoleType)
	}
	if rom.TitleId != titleId {
		t.Errorf("expected TitleId %q, got %q", titleId, rom.TitleId)
	}
}

func TestDSiHeaderDetection(t *testing.T) {
	header := make([]byte, 0x200)
	copy(header[0xC0:], []byte{0x24, 0xFF, 0xAE, 0x51})
	header[0x12] = 0x03
	titleId := "KABC"
	copy(header[0x0C:], []byte(titleId))

	rom, err := fromData(header, "test.dsi")
	if err != nil {
		t.Fatalf("fromData failed: %v", err)
	}
	if rom.ConsoleType != DSi {
		t.Errorf("expected DSi, got %v", rom.ConsoleType)
	}
	if rom.TitleId != titleId {
		t.Errorf("expected TitleId %q, got %q", titleId, rom.TitleId)
	}
}

func TestExtensionFallback(t *testing.T) {
	header := make([]byte, 0x200)

	rom, err := fromData(header, "game.nes")
	if err != nil {
		t.Fatalf("fromData failed: %v", err)
	}
	if rom.ConsoleType != NES {
		t.Errorf("expected NES, got %v", rom.ConsoleType)
	}
}
