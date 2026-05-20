# Cover Sources

How cover art is downloaded from GameTDB and LibRetro.

## GameTDB (NDS/DSi Primary Source)

GameTDB provides NDS boxart organized by region and title ID. Only used for NDS and DSi ROMs.

### URL Pattern

```
https://art.gametdb.com/ds/cover{QUALITY}/{REGION}/{TITLEID}.{EXT}
```

| Quality | Extension | Description |
|---------|-----------|-------------|
| `HQ` | `.jpg` | High quality |
| `M` | `.jpg` | Medium quality |
| `S` | `.png` | Small/standard |

Try in order: HQ → M → S. Use the first successful download.

### Region Mapping

Determine region from the ROM header's `RegionId` byte (offset `0x0F`):

| RegionId | GameTDB Region | Notes |
|----------|---------------|-------|
| `E` | `US` | USA |
| `T` | `US` | USA (alternate) |
| `J` | `JA` | Japan |
| `K` | `KO` | Korea |
| `O` | `EN` | USA/Europe |
| `P` | `EN` | Europe |
| `U` | `EN` | Australia |
| `D` | `DE` | Germany |
| `F` | `FR` | France |
| `H` | `NL` | Netherlands |
| `I` | `IT` | Italy |
| `R` | `RU` | Russia |
| `S` | `ES` | Spain |
| `#` | `HB` | Homebrew |

### Fallback Logic

1. Try the detected region
2. If 404 and region ≠ `EN`, retry with `EN` region
3. If all quality levels fail for both regions → fall through to LibRetro

### DSiWare Placeholder

If a DSi ROM's TitleId starts with `K` or `H` and both GameTDB and LibRetro fail, save the bundled `dsiware.jpg` placeholder image as the cover (processed through the same resize/quantize/BMP pipeline).

## LibRetro (All Systems, NDS Fallback)

LibRetro provides cover thumbnails via GitHub repositories named by console type. Used as the primary source for all non-NDS/DSi systems, and as a fallback for NDS/DSi.

### URL Pattern

```
https://github.com/libretro-thumbnails/{CONSOLE}/raw/master/Named_Boxarts/{NAME}.png
```

### Console String

The console string is derived from the NoIntro description with spaces replaced by underscores:

| Console Type | LibRetro Console String |
|-------------|------------------------|
| NES | `Nintendo_-_Nintendo_Entertainment_System` |
| SNES | `Nintendo_-_Super_Nintendo_Entertainment_System` |
| GB | `Nintendo_-_Game_Boy` |
| GBC | `Nintendo_-_Game_Boy_Color` |
| GBA | `Nintendo_-_Game_Boy_Advance` |
| NDS | `Nintendo_-_Nintendo_DS` |
| DSi | `Nintendo_-_Nintendo_DSi` |
| GameGear | `Sega_-_Game_Gear` |
| Genesis | `Sega_-_Mega_Drive_-_Genesis` |
| MasterSystem | `Sega_-_Master_System_-_Mark_III` |
| FDS | `Nintendo_-_Family_Computer_Disk_System` |

### Name Sanitization

Replace these characters with `_` in the search name:

```
& * / : ` < > ? \ |
```

This matches LibRetro's own sanitization rules (see https://docs.libretro.com/guides/roms-playlists-thumbnails/).

### Download Priority

For each ROM, try names in this order:

1. **NoIntroName** (if available from DB match) — most reliable, uses canonical naming
2. **SearchName** (filename without extension) — fallback when no DB match, or when NoIntroName fails

### Console Type Retry

If the download fails with the ROM's detected `ConsoleType` and `NoIntroConsoleType` differs from `ConsoleType` (and is not Unknown), retry the download using `NoIntroConsoleType` as the console string. This handles cases where a ROM is detected as DSi but its boxart is in the NDS repository.

## General Download Notes

- Use a shared `http.Client` with reasonable timeouts (e.g., 30s)
- Treat HTTP 404 as "not found" (try next option), not as a fatal error
- Treat other HTTP errors / network errors as retries or failures
- Retry failed downloads up to 3 times with exponential backoff (1s, 2s, 4s)
- Skip SSL certificate verification (same as TwilightBoxart — some users have GitHub SSL issues)
