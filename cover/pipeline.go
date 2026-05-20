package cover

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/neonfuz/pico-covers/rom"
)

var ErrNotFound = errors.New("cover not found")

func ProcessCover(r *rom.ROM, outputPath string) error {
	var data []byte
	var err error

	switch r.ConsoleType.BaseType() {
	case rom.NDS, rom.DSi:
		data, err = DownloadGameTDB(r)
	default:
		data, err = DownloadLibRetro(r)
	}

	if err != nil {
		return err
	}

	img, err := imageDecode(data)
	if err != nil {
		return fmt.Errorf("decoding cover: %w", err)
	}

	resized := Resize(img)
	quantized := Quantize(resized)

	return EncodeBMP(quantized, outputPath)
}

func ProcessDSiWarePlaceholder(dsiwarePath string, outputPath string) error {
	data, err := os.ReadFile(dsiwarePath)
	if err != nil {
		return fmt.Errorf("reading dsiware placeholder: %w", err)
	}

	img, err := imageDecode(data)
	if err != nil {
		return fmt.Errorf("decoding dsiware placeholder: %w", err)
	}

	resized := Resize(img)
	quantized := Quantize(resized)

	return EncodeBMP(quantized, outputPath)
}

func imageDecode(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func isNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
