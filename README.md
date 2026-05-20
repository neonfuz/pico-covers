# pico-covers

A CLI tool that downloads boxart for ROM files and saves them as 8bpp BMP covers for [pico-launcher](https://github.com/neonfuz/pico-launcher).

## Usage

```
pico-covers -roms <path> [options]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-roms` | (required) | Directory to scan for ROM files |
| `-covers` | `_pico/covers` | Output directory for cover BMPs |
| `-db` | `~/.cache/pico-covers/nointro.db` | Path to NoIntro database cache |
| `-concurrency` | `runtime.NumCPU()` | Maximum concurrent cover downloads |
| `-refresh-db` | false | Force re-download of NoIntro database |
| `-v` | false | Verbose output |
| `-version` | false | Show version information |

### Examples

```sh
# Basic usage
pico-covers -roms /Volumes/SD/roms

# With custom cover output directory
pico-covers -roms ~/roms -covers /Volumes/SD/_pico/covers

# Refresh database and use verbose mode
pico-covers -roms ~/roms -refresh-db -v
```

## Supported Consoles

| Console | Extensions | Cover Source |
|---------|-----------|-------------|
| NES | `.nes` | LibRetro |
| SNES | `.sfc`, `.smc`, `.snes` | LibRetro |
| Game Boy | `.gb`, `.sgb` | LibRetro |
| Game Boy Color | `.gbc` | LibRetro |
| Game Boy Advance | `.gba` | LibRetro |
| Nintendo DS | `.nds`, `.ds` | GameTDB / LibRetro |
| Nintendo DSi | `.dsi` | GameTDB / LibRetro |
| Game Gear | `.gg` | LibRetro |
| Genesis/Mega Drive | `.gen` | LibRetro |
| Master System | `.sms` | LibRetro |
| Famicom Disk System | `.fds` | LibRetro |
| ZIP archives | `.zip` | Auto-detect inner ROM |

## Output Layout

```
<covers_root>/
├── nds/       # NDS/DSi covers by game code (e.g., ABCD.bmp)
├── gba/       # GBA covers by game code (e.g., ABCE.bmp)
└── user/      # All other covers by filename (e.g., Tetris.gb.bmp)
```

## How It Works

1. Scans the ROM directory for recognized file types
2. Computes SHA1 hash of each ROM and detects console type from headers
3. Looks up game metadata in the NoIntro database
4. Downloads cover art from GameTDB (NDS/DSi) or LibRetro thumbnails (all consoles)
5. Resizes to 106x96 (aspect-ratio preserving), pads to 128x96 black canvas
6. Quantizes to 256 colors with Floyd-Steinberg dithering
7. Saves as 8bpp BMP in the appropriate output directory

## Build

```sh
go build
```

## Database

On first run, pico-covers downloads the NoIntro DAT database from datomatic.no-intro.org and caches it as a gzip-compressed JSON file. Subsequent runs use the cache. Use `-refresh-db` to force a fresh download.
