# Image Processing

How downloaded cover images are resized, quantized, and saved as 8bpp BMP for pico-launcher.

## Target Format

Per the [pico-launcher cover spec](Covers.md):

- **Format**: BMP, 8bpp (256-color indexed)
- **Canvas size**: 128x96 pixels
- **Visible area**: Top-left 106x96 pixels only
- **Padding**: Rightmost 22 pixels (columns 106-127) are ignored by the launcher

## Processing Pipeline

```
1. Download image (JPEG or PNG) via HTTP
2. Decode into image.Image
3. Strip EXIF/metadata
4. Resize to fit within 106x96, preserving aspect ratio (CatmullRom interpolation)
5. Pad to 128x96 canvas with black fill (letterbox/pillarbox as needed)
6. Quantize to 256-color palette (median-cut algorithm)
7. Dither to palette (Floyd-Steinberg)
8. Encode as 8bpp BMP via golang.org/x/image/bmp
```

## Aspect Ratio Handling (Option B)

The cover art is **fitted** into the 106x96 visible area, preserving its original aspect ratio. Black padding fills the unused space.

### Algorithm

1. Calculate the scale factor to fit the source image within 106x96:
   ```
   scaleX = 106.0 / srcWidth
   scaleY = 96.0 / srcHeight
   scale = min(scaleX, scaleY)
   newWidth = int(float64(srcWidth) * scale)
   newHeight = int(float64(srcHeight) * scale)
   ```
2. Resize source to `newWidth x newHeight` using CatmullRom
3. Create a 128x96 black (`image.Black`) canvas
4. Center the resized image on the canvas at offset:
   ```
   offsetX = (106 - newWidth) / 2    // center within visible area
   offsetY = (96 - newHeight) / 2    // center vertically
   ```
5. Draw the resized image onto the canvas at `(offsetX, offsetY)`

The result: cover art is properly proportioned within the visible 106x96 area, with black padding on the sides/top/bottom as needed, and the rightmost 22 columns are already black.

## Resizing

- **Library**: `golang.org/x/image/draw`
- **Interpolation**: `draw.CatmullRom` — high quality, good for downscaling
- **Method**: Use `CatmullRom.Scale()` or create a scaler with `NewScaler()`

## Quantization

- **Library**: `github.com/soniakeys/quant/median`
- **Algorithm**: Median-cut color quantization
- **Target**: 256 colors
- **Input**: `image.Image` (the 128x96 canvas)
- **Output**: `color.Palette` (256 entries)

## Dithering

- **Library**: `golang.org/x/image/draw`
- **Method**: `draw.FloydSteinberg` drawer
- **Input**: Source image + quantized palette
- **Output**: `*image.Paletted` (8bpp indexed image)

The `draw.FloydSteinberg.Draw()` method takes a destination `*image.Paletted`, source `image.Image`, and quantizer. It produces a dithered paletted image that maximizes visual quality with only 256 colors.

## BMP Encoding

- **Library**: `golang.org/x/image/bmp`
- **Method**: `bmp.Encode(writer, img)`
- **Requirement**: Pass `*image.Paletted` — the encoder automatically writes 8bpp indexed BMP with a 256-entry color table
- **Row padding**: The encoder handles 4-byte row alignment automatically

### Expected BMP Structure

| Component | Size |
|-----------|------|
| File header | 14 bytes |
| Info header (BITMAPINFOHEADER) | 40 bytes |
| Color table (256 × 4 bytes) | 1024 bytes |
| Pixel data (128 × 96, padded to 4-byte rows) | 128 × 96 = 12,288 bytes |
| **Total** | ~13,366 bytes |

## DSiWare Placeholder

The bundled `assets/dsiware.jpg` file is processed through the same pipeline (resize → quantize → dither → BMP encode) when used as a fallback cover for DSiWare titles.

## Error Handling

- If image decoding fails (corrupt download) → report error, skip this ROM
- If quantization produces fewer than 256 colors → that's fine, palette is simply smaller
- Never leave a partial/corrupt BMP file — write to a temp file, then rename atomically
