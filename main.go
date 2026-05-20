package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/neonfuz/pico-covers/crawler"
	"github.com/neonfuz/pico-covers/database"
	"github.com/neonfuz/pico-covers/events"
	"github.com/neonfuz/pico-covers/gui"
)

var (
	romsDir     = flag.String("roms", "", "Directory to scan for ROM files (required)")
	coversDir   = flag.String("covers", "_pico/covers", "Output directory for cover BMPs")
	dbPath      = flag.String("db", "", "Path to NoIntro database cache (default ~/.cache/pico-covers/nointro.db)")
	concurrency = flag.Int("concurrency", runtime.NumCPU(), "Maximum concurrent cover downloads")
	refreshDB   = flag.Bool("refresh-db", false, "Force re-download of NoIntro database")
	verbose     = flag.Bool("v", false, "Verbose output")
	showVersion = flag.Bool("version", false, "Show version information")
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "gui" {
		gui.Run()
		return
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  pico-covers -roms <path> [options]\n")
		fmt.Fprintf(os.Stderr, "  pico-covers gui\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	if *romsDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(*romsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: ROMs directory not found: %s\n", *romsDir)
		os.Exit(1)
	}

	ctx := context.Background()
	handler := cliEventHandler(*verbose)

	db := database.NewDatabase()
	if err := db.Initialize(ctx, *dbPath, *refreshDB, handler); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	c := crawler.New(*romsDir, *coversDir, db)
	summary, err := c.Run(ctx, *concurrency, handler)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDone: %d downloaded, %d skipped, %d not found\n",
		summary.Succeeded, summary.Skipped, summary.NotFound)
}

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("pico-covers (unknown version)")
		return
	}
	fmt.Printf("pico-covers %s\n", info.Main.Version)
}

func cliEventHandler(verbose bool) events.EventHandler {
	return func(ev events.ProgressEvent) {
		switch ev.Kind {
		case events.EventDBInit:
			fmt.Println(ev.Detail)
		case events.EventDBLoaded:
			fmt.Printf("Loaded %d ROM records\n", ev.Total)
		case events.EventROMStart:
			if verbose {
				fmt.Printf("[%d/%d] %s\n", ev.Completed, ev.Total, ev.ROMFile)
			}
		case events.EventROMSuccess:
			if verbose {
				if ev.Detail != "" {
					fmt.Printf("  Saved (%s): %s\n", ev.Detail, ev.GameTitle)
				} else {
					fmt.Printf("  Saved: %s\n", ev.GameTitle)
				}
			} else {
				if ev.Detail != "" {
					fmt.Printf("[%d/%d] %s -> %s (%s)\n", ev.Completed, ev.Total, ev.ROMFile, ev.GameTitle, ev.Detail)
				} else {
					fmt.Printf("[%d/%d] %s -> %s\n", ev.Completed, ev.Total, ev.ROMFile, ev.GameTitle)
				}
			}
		case events.EventROMSkipped:
			if verbose {
				fmt.Printf("%s -> %s (already have it)\n", ev.ROMFile, ev.GameTitle)
			} else {
				fmt.Printf("[%d/%d] %s -> %s (already have it)\n", ev.Completed, ev.Total, ev.ROMFile, ev.GameTitle)
			}
		case events.EventROMNotFound:
			if verbose {
				fmt.Printf("%s: no cover found\n", ev.ROMFile)
			} else {
				fmt.Printf("[%d/%d] %s: no match found\n", ev.Completed, ev.Total, ev.ROMFile)
			}
		case events.EventROMError:
			fmt.Printf("%s: parse error: %s\n", ev.ROMFile, ev.Detail)
		case events.EventInfo:
			if verbose {
				fmt.Printf("  %s\n", ev.Detail)
			}
		}
	}
}
