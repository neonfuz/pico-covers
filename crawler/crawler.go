package crawler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/neonfuz/pico-covers/cover"
	"github.com/neonfuz/pico-covers/database"
	"github.com/neonfuz/pico-covers/events"
	"github.com/neonfuz/pico-covers/rom"
)

type Crawler struct {
	RomsDir   string
	CoversDir string
	DB        *database.Database
}

func New(romsDir, coversDir string, db *database.Database) *Crawler {
	return &Crawler{
		RomsDir:   romsDir,
		CoversDir: coversDir,
		DB:        db,
	}
}

type Summary struct {
	Total     int
	Skipped   int
	Succeeded int
	NotFound  int
	Errors    int
}

func (c *Crawler) Run(ctx context.Context, concurrency int, handler events.EventHandler) (*Summary, error) {
	roms, err := c.scanROMs()
	if err != nil {
		return nil, fmt.Errorf("scanning ROMs: %w", err)
	}

	if err := c.ensureDirs(); err != nil {
		return nil, fmt.Errorf("creating output dirs: %w", err)
	}

	handler(events.ProgressEvent{Kind: events.EventInfo, Detail: fmt.Sprintf("Scanning %s...", c.RomsDir)})
	handler(events.ProgressEvent{Kind: events.EventInfo, Detail: fmt.Sprintf("Found %d ROM files", len(roms))})

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	summary := &Summary{Total: len(roms)}

	for i, romFile := range roms {
		select {
		case <-ctx.Done():
			wg.Wait()
			return summary, ctx.Err()
		default:
		}

		wg.Add(1)
		go func(idx int, rf string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := c.processROM(rf, len(roms), idx+1, handler)

			mu.Lock()
			switch result {
			case "ok":
				summary.Succeeded++
			case "skip":
				summary.Skipped++
			case "notfound":
				summary.NotFound++
			case "error":
				summary.Errors++
			}
			mu.Unlock()
		}(i, romFile)
	}

	wg.Wait()

	return summary, nil
}

type romJob struct {
	idx  int
	path string
}

func (c *Crawler) scanROMs() ([]string, error) {
	var roms []string
	err := filepath.WalkDir(c.RomsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := rom.ExtensionMapping[ext]; ok || ext == ".zip" {
			roms = append(roms, path)
		}
		return nil
	})
	return roms, err
}

func (c *Crawler) ensureDirs() error {
	for _, sub := range []string{"nds", "gba", "user"} {
		dir := filepath.Join(c.CoversDir, sub)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (c *Crawler) processROM(romFile string, total, completed int, handler events.EventHandler) string {
	handler(events.ProgressEvent{Kind: events.EventROMStart, ROMFile: filepath.Base(romFile), Total: total, Completed: completed})

	r, err := rom.FromFile(romFile)
	if err != nil {
		handler(events.ProgressEvent{Kind: events.EventROMError, ROMFile: filepath.Base(romFile), Total: total, Completed: completed, Detail: err.Error()})
		return "error"
	}

	c.lookupMetadata(r)

	outputPath := c.outputPath(r)
	if outputPath == "" {
		handler(events.ProgressEvent{Kind: events.EventROMError, ROMFile: r.FileName, Total: total, Completed: completed, Detail: "cannot determine output path"})
		return "error"
	}

	if _, err := os.Stat(outputPath); err == nil {
		handler(events.ProgressEvent{Kind: events.EventROMSkipped, ROMFile: r.FileName, GameTitle: outputPath, Total: total, Completed: completed})
		return "skip"
	}

	handler(events.ProgressEvent{Kind: events.EventInfo, ROMFile: r.FileName, Detail: fmt.Sprintf("SHA1: %s", r.Sha1)})
	if r.NoIntroName != "" {
		handler(events.ProgressEvent{Kind: events.EventInfo, ROMFile: r.FileName, Detail: fmt.Sprintf("DB match: %s (%s)", r.NoIntroName, r.NoIntroConsoleType.String())})
	}
	if r.ConsoleType.BaseType() == rom.NDS || r.ConsoleType.BaseType() == rom.DSi {
		handler(events.ProgressEvent{Kind: events.EventInfo, ROMFile: r.FileName, Detail: "Trying GameTDB..."})
	} else {
		handler(events.ProgressEvent{Kind: events.EventInfo, ROMFile: r.FileName, Detail: fmt.Sprintf("Trying LibRetro: %s", c.libretroDebugURL(r))})
	}

	err = cover.ProcessCover(r, outputPath)

	if err == nil {
		handler(events.ProgressEvent{Kind: events.EventROMSuccess, ROMFile: r.FileName, GameTitle: outputPath, Total: total, Completed: completed})
		return "ok"
	}

	if r.ConsoleType.BaseType() == rom.DSi && isDSiWare(r) {
		placeholderPath, placeholderErr := c.findDSiWarePlaceholder()
		if placeholderErr == nil {
			if perr := cover.ProcessDSiWarePlaceholder(placeholderPath, outputPath); perr == nil {
				handler(events.ProgressEvent{Kind: events.EventROMSuccess, ROMFile: r.FileName, GameTitle: outputPath, Total: total, Completed: completed, Detail: "DSiWare placeholder"})
				return "ok"
			}
		}
	}

	handler(events.ProgressEvent{Kind: events.EventROMNotFound, ROMFile: r.FileName, Detail: "no cover found", Total: total, Completed: completed})
	return "notfound"
}

func (c *Crawler) lookupMetadata(r *rom.ROM) {
	meta := c.DB.FindBySHA1(r.Sha1)
	if meta != nil {
		r.NoIntroName = meta.Name
		r.NoIntroConsoleType = rom.ConsoleType(meta.ConsoleSubType)
		return
	}

	if r.TitleId != "" {
		dbType := database.ConsoleType(r.ConsoleType)
		meta = c.DB.FindByTitleID(r.TitleId, dbType)
		if meta != nil {
			r.NoIntroName = meta.Name
			r.NoIntroConsoleType = rom.ConsoleType(meta.ConsoleSubType)
		}
	}
}

func (c *Crawler) outputPath(r *rom.ROM) string {
	switch r.ConsoleType.BaseType() {
	case rom.NDS, rom.DSi:
		if r.TitleId != "" {
			return filepath.Join(c.CoversDir, "nds", r.TitleId+".bmp")
		}
		return filepath.Join(c.CoversDir, "user", r.FileName+".bmp")
	case rom.GBA:
		if r.TitleId != "" {
			return filepath.Join(c.CoversDir, "gba", r.TitleId+".bmp")
		}
		return filepath.Join(c.CoversDir, "user", r.FileName+".bmp")
	default:
		return filepath.Join(c.CoversDir, "user", r.FileName+".bmp")
	}
}

func (c *Crawler) libretroDebugURL(r *rom.ROM) string {
	consoleStr := r.ConsoleType.LibRetroConsoleString()
	name := r.NoIntroName
	if name == "" {
		name = r.SearchName
	}
	sanitized := strings.NewReplacer(
		"&", "_", "*", "_", "/", "_", ":", "_",
		"<", "_", ">", "_", "?", "_", "\\", "_",
		"|", "_", " ", "_",
	).Replace(name)
	return fmt.Sprintf("https://github.com/libretro-thumbnails/%s/raw/master/Named_Boxarts/%s.png",
		consoleStr, sanitized)
}

func isDSiWare(r *rom.ROM) bool {
	if len(r.TitleId) == 0 {
		return false
	}
	return r.TitleId[0] == 'K' || r.TitleId[0] == 'H'
}

func (c *Crawler) findDSiWarePlaceholder() (string, error) {
	candidates := []string{
		"assets/dsiware.jpg",
		filepath.Join(filepath.Dir(os.Args[0]), "assets/dsiware.jpg"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("dsiware placeholder not found")
}
