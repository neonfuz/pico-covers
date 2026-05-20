package rom

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrNoMatch = errors.New("no cover found")

type ROM struct {
	FileName           string
	SearchName         string
	Sha1               string
	ConsoleType        ConsoleType
	TitleId            string
	RegionId           byte
	NoIntroName        string
	NoIntroConsoleType ConsoleType
	header             []byte
}

type ROMOption func(*ROM)

func newROM(fileName string, consoleType ConsoleType, opts ...ROMOption) *ROM {
	ext := filepath.Ext(fileName)
	searchName := strings.TrimSuffix(fileName, ext)
	r := &ROM{
		FileName:           fileName,
		SearchName:         searchName,
		ConsoleType:        consoleType,
		NoIntroName:        "",
		NoIntroConsoleType: Unknown,
		RegionId:           0,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithHeader(header []byte) ROMOption {
	return func(r *ROM) {
		r.header = header
	}
}

func WithSha1(sha1Hex string) ROMOption {
	return func(r *ROM) {
		r.Sha1 = sha1Hex
	}
}

func ComputeSHA1(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}

func HeaderMatch(header []byte, offset int, pattern ...byte) bool {
	if len(header) < offset+len(pattern) {
		return false
	}
	for i, b := range pattern {
		if header[offset+i] != b {
			return false
		}
	}
	return true
}

func readString(header []byte, offset, length int) string {
	end := offset + length
	if end > len(header) {
		end = len(header)
	}
	buf := header[offset:end]
	nullIdx := bytes.IndexByte(buf, 0)
	if nullIdx >= 0 {
		buf = buf[:nullIdx]
	}
	return string(buf)
}

func FromFile(filePath string) (*ROM, error) {
	fileName := filepath.Base(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".zip" {
		return fromZip(data, fileName)
	}

	return fromData(data, fileName)
}

func fromZip(data []byte, zipFileName string) (*ROM, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	for _, f := range zipReader.File {
		ext := strings.ToLower(filepath.Ext(f.Name))
		if _, ok := ExtensionMapping[ext]; ok {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			innerData, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				return nil, err
			}
			return fromData(innerData, f.Name)
		}
	}

	ext := strings.ToLower(filepath.Ext(zipFileName))
	consoleType := ExtensionMapping[ext]

	rom := newROM(zipFileName, consoleType, WithSha1(ComputeSHA1(data)))
	if consoleType == Unknown {
		return fromUnknown(data, rom), nil
	}
	return rom, nil
}

func fromData(data []byte, fileName string) (*ROM, error) {
	sha1Hex := ComputeSHA1(data)

	headerSize := 328
	if len(data) < headerSize {
		headerSize = len(data)
	}
	header := data[:headerSize]

	rom := detectROM(fileName, header, sha1Hex)
	return rom, nil
}

func detectROM(fileName string, header []byte, sha1Hex string) *ROM {
	ext := strings.ToLower(filepath.Ext(fileName))

	if HeaderMatch(header, 0xC0, 0x24, 0xFF, 0xAE, 0x51) {
		if len(header) > 0x12 && header[0x12] == 0x03 {
			return newDSiROM(fileName, header, sha1Hex)
		}
		return newNDSROM(fileName, header, sha1Hex)
	}

	if HeaderMatch(header, 0x04, 0x24, 0xFF, 0xAE, 0x51) {
		return newGBAROM(fileName, header, sha1Hex)
	}

	if HeaderMatch(header, 0x104, 0xCE, 0xED, 0x66, 0x66) ||
		HeaderMatch(header, 0x100, 0x00, 0xC3, 0x50, 0x01) ||
		HeaderMatch(header, 0x104, 0x11, 0x23, 0xF1, 0x1E) {
		if len(header) > 0x143 {
			switch header[0x143] {
			case 0x80, 0xC0:
				return newGBCROM(fileName, header, sha1Hex)
			}
		}
		return newGBROM(fileName, header, sha1Hex)
	}

	if consoleType, ok := ExtensionMapping[ext]; ok {
		rom := newROM(fileName, consoleType, WithHeader(header), WithSha1(sha1Hex))
		return rom
	}

	return newROM(fileName, Unknown, WithHeader(header), WithSha1(sha1Hex))
}
