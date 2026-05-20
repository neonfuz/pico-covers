# ROM Parsing

How ROM files are identified, parsed, and matched to console types.

## File Discovery

- Recursively enumerate all files in the ROM directory
- Filter by file extension using the extension mapping below
- Skip files where a cover already exists at the target output path

## Extension Mapping

| Extension | Console Type | LibRetro Console String |
|-----------|-------------|------------------------|
| `.nes` | NES | `Nintendo_-_Nintendo_Entertainment_System` |
| `.sfc` | SNES | `Nintendo_-_Super_Nintendo_Entertainment_System` |
| `.smc` | SNES | `Nintendo_-_Super_Nintendo_Entertainment_System` |
| `.snes` | SNES | `Nintendo_-_Super_Nintendo_Entertainment_System` |
| `.gb` | GB | `Nintendo_-_Game_Boy` |
| `.sgb` | GB | `Nintendo_-_Game_Boy` |
| `.gbc` | GBC | `Nintendo_-_Game_Boy_Color` |
| `.gba` | GBA | `Nintendo_-_Game_Boy_Advance` |
| `.nds` | NDS | `Nintendo_-_Nintendo_DS` |
| `.ds` | NDS | `Nintendo_-_Nintendo_DS` |
| `.dsi` | DSi | `Nintendo_-_Nintendo_DSi` |
| `.gg` | GameGear | `Sega_-_Game_Gear` |
| `.gen` | Genesis | `Sega_-_Mega_Drive_-_Genesis` |
| `.sms` | MasterSystem | `Sega_-_Master_System_-_Mark_III` |
| `.fds` | FDS | `Nintendo_-_Family_Computer_Disk_System` |
| `.zip` | Auto-detect inner | (varies) |

## SHA1 Hashing

- Compute SHA1 of the **entire file contents**
- For `.zip` files: extract the first entry with a recognized extension, decompress to memory, then hash the decompressed data
- Store as lowercase hex string

## Header-Based Type Detection

Read the first 328 bytes of the ROM file as a header. Apply byte-pattern matching at specific offsets:

| Offset | Bytes (hex) | Detection |
|--------|-------------|-----------|
| `0x104` | `CE ED 66 66` | GB/GBC (Nintendo logo) |
| `0x100` | `00 C3 50 01` | GB/GBC (alternate header) |
| `0x104` | `11 23 F1 1E` | GB/GBC (another variant) |
| `0x04` | `24 FF AE 51` | GBA (Nintendo logo at offset 4) |
| `0xC0` | `24 FF AE 51` | NDS or DSi (Nintendo logo at offset 192) |

### GB vs GBC

If a GB-family header is detected, check byte at offset `0x143`:
- `0x80` → GBC (GBC compatible)
- `0xC0` → GBC (GBC only)
- Anything else → GB

### NDS vs DSi

If the NDS/DSi header is detected, check byte at offset `0x012`:
- `0x03` → DSi
- Anything else → NDS

## Header Field Extraction

### NDS/DSi

| Field | Offset | Length | Description |
|-------|--------|--------|-------------|
| Title | `0x00` | 12 | Game title (ASCII) |
| TitleId | `0x0C` | 4 | Game code (e.g., "ABCD") |
| RegionId | `0x0F` | 1 | Region character (e.g., 'E' = USA) |

### GBA

| Field | Offset | Length | Description |
|-------|--------|--------|-------------|
| TitleId | `0xAC` | 4 | Game code (e.g., "ABCE") |

### GB/GBC

| Field | Offset | Length | Description |
|-------|--------|--------|-------------|
| Title | `0x134` | 15 (GBC) / 11 (GB) | Game title |
| TitleId | `0x13F` | 4 (GBC only, when byte 0x143 == 0xC0) | Manufacturer code |

Note: GB/GBC title IDs are manufacturer codes, not the same format as NDS/GBA game codes. They are not used for GameTDB lookups.

## Fallback: Extension-Only Detection

If no header pattern matches but the file extension is recognized, create an "Unknown" ROM with the console type from the extension mapping. These ROMs support SHA1 and filename matching only — no title ID extraction is possible.

## ROM Data Model

Each parsed ROM should expose:

- `FileName` — original filename (with extension)
- `SearchName` — filename without extension (used for LibRetro filename matching)
- `Sha1` — lowercase hex SHA1 of file contents
- `ConsoleType` — detected console type
- `TitleId` — extracted game code (empty if not available)
- `RegionId` — region character (NDS/DSi only)
- `NoIntroName` — populated from DB lookup (empty if no match)
- `NoIntroConsoleType` — populated from DB lookup (Unknown if no match)

## NoIntro Console Type Mapping

Some console types in the NoIntro DB map to a base type for deduplication:

| NoIntro Type | Maps To |
|-------------|---------|
| Nintendo DS | Nintendo DS |
| Nintendo DS (Download Play) | Nintendo DS |
| Nintendo DSi | Nintendo DSi |
| Nintendo DSi (Digital) | Nintendo DSi |
