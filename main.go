package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/neonfuz/pico-covers/internal/crawler"
	"github.com/neonfuz/pico-covers/internal/database"
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
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	if *romsDir == "" {
		fmt.Fprintln(os.Stderr, "Error: -roms flag is required")
		fmt.Fprintln(os.Stderr, "Usage: pico-covers -roms <path> [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if _, err := os.Stat(*romsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: ROMs directory not found: %s\n", *romsDir)
		os.Exit(1)
	}

	ctx := context.Background()

	db := database.NewDatabase()
	if err := db.Initialize(ctx, *dbPath, *refreshDB); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d ROM records\n", len(db.Records))

	c := crawler.New(*romsDir, *coversDir, db)
	summary, err := c.Run(ctx, *concurrency, *verbose)
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
