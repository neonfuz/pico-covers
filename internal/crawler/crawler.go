package crawler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/neonfuz/pico-covers/internal/cover"
	"github.com/neonfuz/pico-covers/internal/database"
	"github.com/neonfuz/pico-covers/internal/rom"
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

func (c *Crawler) Run(ctx context.Context, concurrency int, verbose bool) (*Summary, error) {
	roms, err := c.scanROMs()
	if err != nil {
		return nil, fmt.Errorf("scanning ROMs: %w", err)
	}

	if err := c.ensureDirs(); err != nil {
		return nil, fmt.Errorf("creating output dirs: %w", err)
	}

	fmt.Printf("Scanning %s...\n", c.RomsDir)
	fmt.Printf("Found %d ROM files\n", len(roms))

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	summary := &Summary{Total: len(roms)}

	for i, romFile := range roms {
		wg.Add(1)
		go func(idx int, rf string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := c.processROM(rf, verbose)

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

func (c *Crawler) processROM(romFile string, verbose bool) string {
	r, err := rom.FromFile(romFile)
	if err != nil {
		fmt.Printf("%s: parse error: %v\n", filepath.Base(romFile), err)
		return "error"
	}

	c.lookupMetadata(r)

	outputPath := c.outputPath(r)
	if outputPath == "" {
		fmt.Printf("%s: cannot determine output path\n", r.FileName)
		return "error"
	}

	if _, err := os.Stat(outputPath); err == nil {
		if verbose {
			fmt.Printf("%s -> %s (already have it)\n", r.FileName, outputPath)
		}
		return "skip"
	}

	if verbose {
		fmt.Printf("%s\n", r.FileName)
		fmt.Printf("  SHA1: %s\n", r.Sha1)
		if r.NoIntroName != "" {
			fmt.Printf("  DB match: %s (%s)\n", r.NoIntroName, r.NoIntroConsoleType.String())
		}
		if r.ConsoleType.BaseType() == rom.NDS || r.ConsoleType.BaseType() == rom.DSi {
			fmt.Printf("  Trying GameTDB...\n")
		} else {
			fmt.Printf("  Trying LibRetro: %s\n", c.libretroDebugURL(r))
		}
	}

	err = cover.ProcessCover(r, outputPath)

	if err == nil {
		if verbose {
			fmt.Printf("  Saved: %s\n", outputPath)
		} else {
			fmt.Printf("%s -> %s\n", r.FileName, outputPath)
		}
		return "ok"
	}

	if r.ConsoleType.BaseType() == rom.DSi && isDSiWare(r) {
		placeholderPath, placeholderErr := c.findDSiWarePlaceholder()
		if placeholderErr == nil {
			if perr := cover.ProcessDSiWarePlaceholder(placeholderPath, outputPath); perr == nil {
				if verbose {
					fmt.Printf("  Saved (DSiWare placeholder): %s\n", outputPath)
				} else {
					fmt.Printf("%s -> %s (DSiWare placeholder)\n", r.FileName, outputPath)
				}
				return "ok"
			}
		}
	}

	fmt.Printf("%s: no cover found\n", r.FileName)
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
