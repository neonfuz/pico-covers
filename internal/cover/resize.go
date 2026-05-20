package cover

import (
	"image"

	"golang.org/x/image/draw"
)

func Resize(src image.Image) *image.RGBA {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	targetMaxW := 106
	targetMaxH := 96

	scale := float64(targetMaxW) / float64(srcW)
	if float64(targetMaxH)/float64(srcH) < scale {
		scale = float64(targetMaxH) / float64(srcH)
	}

	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)

	canvas := image.NewRGBA(image.Rect(0, 0, 128, 96))

	offsetX := (targetMaxW - newW) / 2
	offsetY := (targetMaxH - newH) / 2

	dstRect := image.Rect(offsetX, offsetY, offsetX+newW, offsetY+newH)
	draw.CatmullRom.Scale(canvas, dstRect, src, srcBounds, draw.Over, nil)

	return canvas
}
