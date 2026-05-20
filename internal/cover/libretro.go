package cover

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/neonfuz/pico-covers/internal/rom"
)

var sanitizeReplacer = strings.NewReplacer(
	"&", "_", "*", "_", "/", "_", ":", "_",
	"<", "_", ">", "_", "?", "_", "\\", "_",
	"|", "_", " ", "_",
)

func sanitizeName(name string) string {
	return sanitizeReplacer.Replace(name)
}

func DownloadLibRetro(r *rom.ROM) ([]byte, error) {
	names := make([]string, 0, 2)
	if r.NoIntroName != "" && r.NoIntroName != r.SearchName {
		names = append(names, r.NoIntroName)
	}
	names = append(names, r.SearchName)

	consoleTypes := make([]rom.ConsoleType, 0, 2)
	consoleTypes = append(consoleTypes, r.ConsoleType)
	if r.NoIntroConsoleType != rom.Unknown && r.NoIntroConsoleType != r.ConsoleType {
		consoleTypes = append(consoleTypes, r.NoIntroConsoleType)
	}

	for _, name := range names {
		sanitized := sanitizeName(name)
		for _, ct := range consoleTypes {
			consoleStr := ct.LibRetroConsoleString()
			if consoleStr == "" {
				continue
			}
			rawURL := fmt.Sprintf("https://github.com/libretro-thumbnails/%s/raw/master/Named_Boxarts/%s.png",
				url.PathEscape(consoleStr), url.PathEscape(sanitized))
			data, err := Download(rawURL)
			if err == nil {
				return data, nil
			}
		}
	}

	return nil, ErrNotFound
}
