# CLI Interface

Command-line interface design for pico-covers.

## Usage

```
pico-covers -roms <path> [options]
pico-covers gui
```

The `gui` subcommand launches a Fyne-based graphical interface. See [GUI.md](GUI.md).

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-roms` | string | (required) | Directory to scan for ROM files |
| `-covers` | string | `/_pico/covers` | Output directory for cover BMPs |
| `-db` | string | `~/.cache/pico-covers/nointro.db` | Path to NoIntro database cache |
| `-concurrency` | int | `runtime.NumCPU()` | Maximum concurrent cover downloads |
| `-refresh-db` | bool | false | Force re-download of NoIntro database |
| `-v` | bool | false | Verbose output (show each download URL) |

## Execution Flow

1. Parse flags
2. Initialize NoIntro database (download if needed, or load from cache)
3. Scan ROM directory recursively
4. For each ROM (concurrent, bounded by `-concurrency`):
   a. Check if cover already exists → skip
   b. Parse ROM header → detect type, extract TitleId, compute SHA1
   c. Look up metadata in NoIntro DB (SHA1 first, then TitleId)
   d. Download cover art (GameTDB for NDS/DSi, LibRetro for all)
   e. Resize, quantize, dither, encode as 8bpp BMP
   f. Save to appropriate output directory
5. Print summary

## Output Format

Default (non-verbose):
```
Scanning /path/to/roms...
Downloading NoIntro database... (only on first run or -refresh-db)
Loaded 145,232 ROM records

[1/50] Pokemon - Emerald Version.gba → covers/gba/BPRE.bmp ✓
[2/50] Tetris.gb → covers/user/Tetris.gb.bmp ✓
[3/50] Unknown_game.nes → no match found
[4/50] Mario Kart DS.nds → covers/nds/AMKE.bmp ✓ (already have it)

Done: 2 downloaded, 1 skipped, 1 not found
```

Verbose (`-v`):
```
[1/50] Pokemon - Emerald Version.gba
  SHA1: abc123...
  DB match: "Pokemon - Emerald Version" (GBA)
  Trying LibRetro: https://github.com/libretro-thumbnails/Nintendo_-_Game_Boy_Advance/raw/master/Named_Boxarts/Pokemon_-_Emerald_Version.png
  Downloaded, resizing to 106x96 → 128x96 canvas, quantizing 256 colors
  Saved: covers/gba/BPRE.bmp
```

## Event System

The crawler emits typed events through an `EventHandler` callback rather than printing directly. The CLI implements a handler that formats events for stdout. The same events drive the [GUI](GUI.md) progress view. This means all output (CLI verbose, CLI non-verbose, GUI) flows through the same event stream.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (some ROMs may not have found covers, that's OK) |
| 1 | Fatal error (missing required flag, ROM directory not found, etc.) |
