# Output Layout

How cover files are named and organized on disk for pico-launcher compatibility.

## Directory Structure

```
<covers_root>/
├── nds/
│   ├── ABCD.bmp
│   ├── AMHE.bmp
│   └── ...
├── gba/
│   ├── ABCE.bmp
│   ├── BPRE.bmp
│   └── ...
└── user/
    ├── Pokemon - Emerald Version.gba.bmp
    ├── Tetris.gb.bmp
    ├── Super Mario Bros..nes.bmp
    └── ...
```

## File Placement Rules

Covers are placed in subfolders of the `/_pico/covers` directory according to these rules:

### `nds/` — NDS/DSi covers by game code

- **Eligible**: NDS and DSi ROMs that have a TitleId extracted from their header
- **Filename**: `{TitleId}.bmp` (e.g., `ABCD.bmp`)
- The TitleId is the 4-character game code from offset `0x0C` of the NDS/DSi header

### `gba/` — GBA covers by game code

- **Eligible**: GBA ROMs that have a TitleId extracted from their header
- **Filename**: `{TitleId}.bmp` (e.g., `ABCE.bmp`)
- The TitleId is the 4-character game code from offset `0xAC` of the GBA header

### `user/` — Covers by filename (all other systems)

- **Eligible**: All ROMs without a title ID, or ROMs of systems other than NDS/DSi/GBA
- **Filename**: `{rom_filename}.bmp` (e.g., `Tetris.gb.bmp`, `Super Mario Bros..nes.bmp`)
- The full original ROM filename (with extension) is used, with `.bmp` appended
- This folder is suited for GB, GBC, NES, SNES, Genesis, SMS, GG, FDS, and any other system

## Precedence

Per the pico-launcher spec: a `user/` cover takes precedence over a game-code-based cover. However, we do not write both — we write to the appropriate directory based on the ROM type.

## Skip Logic

Before downloading, check if a cover already exists at the target output path:
- For NDS/DSi: check `covers/nds/{TitleId}.bmp`
- For GBA: check `covers/gba/{TitleId}.bmp`
- For others: check `covers/user/{filename}.bmp`

If the file exists, skip the ROM and report "already have it".

## Directory Creation

Create output subdirectories as needed (`nds/`, `gba/`, `user/`) if they don't exist.

## Atomic Writes

Write the BMP to a temporary file in the target directory first, then rename to the final path. This avoids leaving corrupt partial files if the process is interrupted.
