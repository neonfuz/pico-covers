package cover

import (
	"image"
	"os"
	"path/filepath"

	"golang.org/x/image/bmp"
)

func EncodeBMP(img *image.Paletted, filePath string) error {
	dir := filepath.Dir(filePath)

	tmpFile, err := os.CreateTemp(dir, "cover-*.bmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	if err := bmp.Encode(tmpFile, img); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}
