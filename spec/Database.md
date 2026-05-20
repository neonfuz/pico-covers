# NoIntro Database

How the NoIntro database is downloaded, parsed, cached, and queried for ROM metadata.

## Overview

The NoIntro database provides a mapping from SHA1 hashes and title IDs/serials to game names. This is essential for finding the correct cover art on LibRetro, which uses NoIntro naming conventions for its thumbnail repositories.

## Download

### NoIntro DAT Files

1. Send GET request to `https://datomatic.no-intro.org/index.php?page=download&fun=wut` to establish a session cookie
2. POST to the same URL with form data:
   - `download=Download`
   - `sel_s=<ConsoleType>` (e.g., `Nintendo - Nintendo DS`)
3. Response is a ZIP file containing an XML DAT file
4. Extract and parse the XML

Download a DAT for each console type. Skip `Unknown`.

### LibRetro DAT Files

Supplementary SHA1 sources. Download from:

| Console | URL |
|---------|-----|
| NES | `https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Nintendo%20Entertainment%20System.dat` |

These use a custom non-standard format (see LibRetro DAT Parsing below).

## NoIntro XML Format

The XML structure follows the NoIntro DAT specification:

```xml
<datafile>
  <game name="Game Name" game_id="XXXX">
    <rom sha1="abc123..." serial="ABCD" status="verified"/>
  </game>
  ...
</datafile>
```

### Fields Extracted Per Game

| XML Attribute | DB Field | Notes |
|--------------|----------|-------|
| `game@name` | `Name` | NoIntro canonical name |
| `game@game_id` | `GameId` | |
| `rom@sha1` | `Sha1` | Lowercased for comparison |
| `rom@serial` | `Serial` | Used for title ID matching |
| `rom@status` | `Status` | Used to filter "bad" dumps |

### Console Type Merging

When storing records, map the console type to its base type:

- `Nintendo DS (Download Play)` → `Nintendo DS`
- `Nintendo DSi (Digital)` → `Nintendo DSi`

The original type is preserved as `ConsoleSubType` for LibRetro retry logic.

## LibRetro DAT Parsing

LibRetro uses a custom parenthesized key-value format (not XML). Example:

```
game (
  name "Game Name"
  rom ( name "filename.nes" size 262144 crc ABCD1234 sha1 abc123... )
)
```

Parse this format to extract:
- `name` → `Name`
- `rom>sha1` → `Sha1` (lowercased)

## Caching

### Local Cache Format

- Gzip-compressed JSON file
- Contains an array of `RomMetaData` objects
- Default path: `~/.cache/pico-covers/nointro.db`

### Cache Behavior

- If cache file exists and is valid → load from cache, skip all downloads
- If cache is missing or corrupt → download all DATs, build cache, save
- `-refresh-db` flag forces re-download even if cache exists

### RomMetaData Model

```
RomMetaData {
    ConsoleType    ConsoleType    // Base type (merged)
    ConsoleSubType ConsoleType    // Original type from DAT
    Name           string         // NoIntro game name
    Sha1           string         // Lowercase hex SHA1
    Serial         string         // From DAT (for title ID matching)
    GameId         string         // From DAT
    Status         string         // "verified", "bad", etc.
}
```

## Lookup Queries

### SHA1 Lookup

- Search all records for a case-insensitive SHA1 match
- This is the most reliable matching method
- Works for all console types

### Title ID / Serial Lookup

- Only if SHA1 lookup fails
- Only for ROMs that have a TitleId extracted from their header
- Search records where:
  - `ConsoleType` matches the ROM's console type, AND
  - `Serial` matches `TitleId` (case-insensitive) OR `GameId` matches `TitleId`
- If multiple matches: prefer records where `Status` does not contain "bad"
- Primary method for NDS/DSi games

### Result

When a match is found, populate the ROM's:
- `NoIntroName` → from matched record's `Name`
- `NoIntroConsoleType` → from matched record's `ConsoleSubType`

These are used by the cover download pipeline to construct LibRetro URLs.

## SHA1 Index Optimization

For fast SHA1 lookups across potentially millions of records, build an in-memory map:

```
map[string][]int  // sha1_lowercase → indices into metadata slice
```

This avoids O(n) linear scans on every lookup.
