# Implementation Plan: pico-covers

A CLI tool written in Go that downloads boxart for ROM files and saves them as 8bpp BMP covers for pico-launcher.

## Spec Documents

- [Cover format](spec/Covers.md) — pico-launcher BMP requirements
- [ROM parsing](spec/ROM-Parsing.md) — header detection, SHA1, type identification
- [Database](spec/Database.md) — NoIntro DB download, cache, lookup
- [Cover sources](spec/Cover-Sources.md) — GameTDB and LibRetro download logic
- [Image processing](spec/Image-Processing.md) — resize, quantize, dither, BMP encode
- [Output layout](spec/Output-Layout.md) — file naming and directory structure
- [CLI](spec/CLI.md) — command-line interface design

## Phase 1: Project Scaffolding & ROM Parsing

- [ ] Initialize Go module (`go mod init github.com/user/pico-covers`)
- [ ] Add dependencies: `golang.org/x/image`, `github.com/soniakeys/quant/median`
- [ ] Create `internal/rom/rom.go` — base ROM struct, `FromFile()`, `FromStream()`, SHA1 computation, header reading, byte matching utility
- [ ] Create `internal/rom/nds.go` — NDS ROM type: TitleId (offset 0x0C, len 4), RegionId (offset 0x0F), region mapping per [Cover Sources](spec/Cover-Sources.md)
- [ ] Create `internal/rom/dsi.go` — DSi ROM type: same as NDS, detected by byte 0x12 == 0x03
- [ ] Create `internal/rom/gba.go` — GBA ROM type: TitleId (offset 0xAC, len 4)
- [ ] Create `internal/rom/gb.go` — GB ROM type: no title ID, header detection at offset 0x104/0x100
- [ ] Create `internal/rom/gbc.go` — GBC ROM type: detected by byte 0x143 == 0x80 or 0xC0
- [ ] Create `internal/rom/unknown.go` — extension-only fallback ROM type
- [ ] Create `internal/rom/console.go` — ConsoleType enum, extension mapping, LibRetro console strings per [ROM Parsing](spec/ROM-Parsing.md)
- [ ] Add `.zip` support in `FromFile()` — extract first entry with recognized extension per [ROM Parsing](spec/ROM-Parsing.md)
- [ ] Write tests for ROM header detection using small test ROM headers

## Phase 2: NoIntro Database

- [ ] Create `internal/database/metadata.go` — RomMetaData struct, ConsoleType enum per [Database](spec/Database.md)
- [ ] Create `internal/database/nointro.go` — download NoIntro DAT XMLs from datomatic.no-intro.org (GET session + POST per console type), parse XML, extract game/rom fields per [Database](spec/Database.md)
- [ ] Create `internal/database/libretro_dat.go` — download and parse LibRetro .dat format (custom parenthesized key-value) per [Database](spec/Database.md)
- [ ] Create `internal/database/database.go` — Database struct with Initialize(), load/save gzip JSON cache, SHA1 index map per [Database](spec/Database.md)
- [ ] Implement `FindBySHA1(sha1)` — case-insensitive lookup using prebuilt SHA1 index map
- [ ] Implement `FindByTitleID(titleID, consoleType)` — search by serial/gameId with "bad" dump filtering per [Database](spec/Database.md)
- [ ] Implement console type merging (DS Download Play → NDS, DSi Digital → DSi) per [ROM Parsing](spec/ROM-Parsing.md)
- [ ] Write tests for DB lookup logic with in-memory test data

## Phase 3: Cover Download & Image Processing

- [ ] Create `internal/cover/downloader.go` — shared HTTP client, download function with retry (3 attempts, exponential backoff) per [Cover Sources](spec/Cover-Sources.md)
- [ ] Create `internal/cover/resize.go` — fit image into 106x96 preserving aspect ratio using `golang.org/x/image/draw.CatmullRom`, pad to 128x96 black canvas per [Image Processing](spec/Image-Processing.md)
- [ ] Create `internal/cover/quantize.go` — median-cut quantization to 256 colors via `github.com/soniakeys/quant/median`, Floyd-Steinberg dithering via `golang.org/x/image/draw.FloydSteinberg`, produce `*image.Paletted` per [Image Processing](spec/Image-Processing.md)
- [ ] Create `internal/cover/encode.go` — encode `*image.Paletted` as 8bpp BMP via `golang.org/x/image/bmp`, atomic write (temp file + rename) per [Image Processing](spec/Image-Processing.md)
- [ ] Create `internal/cover/gametdb.go` — GameTDB URL construction, quality fallback (HQ→M→S), region detection from header, EN region fallback per [Cover Sources](spec/Cover-Sources.md)
- [ ] Create `internal/cover/libretro.go` — LibRetro URL construction, console string mapping, name sanitization (`&*/:`<>?\|` → `_`), NoIntroName/SearchName priority, console type retry per [Cover Sources](spec/Cover-Sources.md)
- [ ] Create `internal/cover/pipeline.go` — full pipeline: download → decode → strip metadata → resize → quantize → dither → encode BMP → save per [Image Processing](spec/Image-Processing.md)
- [ ] Copy `TwilightBoxart/img/dsiware.jpg` to `assets/dsiware.jpg` — DSiWare placeholder per [Cover Sources](spec/Cover-Sources.md)
- [ ] Write tests for image processing pipeline with a small test image

## Phase 4: Crawler & Output

- [ ] Create `internal/crawler/crawler.go` — main orchestrator per [Output Layout](spec/Output-Layout.md)
- [ ] Implement ROM directory scanning (recursive, filtered by extension mapping)
- [ ] Implement skip logic — check if cover already exists at target path per [Output Layout](spec/Output-Layout.md)
- [ ] Implement output path determination:
  - NDS/DSi with TitleId → `covers/nds/{TitleId}.bmp`
  - GBA with TitleId → `covers/gba/{TitleId}.bmp`
  - All others → `covers/user/{romfilename}.bmp`
  per [Output Layout](spec/Output-Layout.md)
- [ ] Implement directory creation for output subdirs (`nds/`, `gba/`, `user/`)
- [ ] Implement concurrent processing with semaphore (bounded by `-concurrency`) — parse ROM + DB lookup (sequential), download + process (concurrent) per [CLI](spec/CLI.md)
- [ ] Implement progress reporting per [CLI](spec/CLI.md)
- [ ] Implement summary output (downloaded / skipped / not found counts) per [CLI](spec/CLI.md)

## Phase 5: CLI & Integration

- [ ] Create `main.go` — parse flags, validate required args, call crawler per [CLI](spec/CLI.md)
- [ ] Implement flag parsing: `-roms`, `-covers`, `-db`, `-concurrency`, `-refresh-db`, `-v` per [CLI](spec/CLI.md]
- [ ] Implement default values for flags (`-covers`, `-db`, `-concurrency`) per [CLI](spec/CLI.md)
- [ ] Implement exit codes (0 success, 1 fatal error) per [CLI](spec/CLI.md)
- [ ] Add version flag (`-version`) with Go build info
- [ ] End-to-end test: run against a small directory of test ROMs, verify BMP output

## Phase 6: Polish

- [ ] Add `go vet` and `golangci-lint` passes
- [ ] Ensure all errors are handled (no silent failures)
- [ ] Verify BMP output is valid 8bpp with correct dimensions (128x96) using an image viewer
- [ ] Test on a real pico-launcher SD card layout
- [ ] Write README.md with usage instructions
