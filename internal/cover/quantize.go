package cover

import (
	"image"
	"image/color"

	"github.com/soniakeys/quant/median"
	"golang.org/x/image/draw"
)

func Quantize(img image.Image) *image.Paletted {
	q := median.Quantizer(256)
	palette := q.Quantize(make(color.Palette, 0, 256), img)
	dst := image.NewPaletted(img.Bounds(), palette)
	draw.FloydSteinberg.Draw(dst, img.Bounds(), img, image.Point{})
	return dst
}
