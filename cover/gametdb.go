package cover

import (
	"fmt"

	"github.com/neonfuz/pico-covers/rom"
)

var regionMap = map[byte]string{
	'E': "US", 'T': "US",
	'J': "JA", 'K': "KO",
	'O': "EN", 'P': "EN", 'U': "EN",
	'D': "DE", 'F': "FR", 'H': "NL",
	'I': "IT", 'R': "RU", 'S': "ES",
	'#': "HB",
}

type qualityEntry struct {
	quality string
	ext     string
}

var qualities = []qualityEntry{
	{"HQ", ".jpg"},
	{"M", ".jpg"},
	{"S", ".png"},
}

func DownloadGameTDB(rom *rom.ROM) ([]byte, error) {
	region := regionMap[rom.RegionId]
	if region == "" {
		region = "EN"
	}

	titleID := rom.TitleId
	if titleID == "" {
		return nil, fmt.Errorf("%w: no TitleId for GameTDB", ErrNotFound)
	}

	for _, q := range qualities {
		url := fmt.Sprintf("https://art.gametdb.com/ds/cover%s/%s/%s%s",
			q.quality, region, titleID, q.ext)
		data, err := Download(url)
		if err == nil {
			return data, nil
		}
		if !isNotFound(err) {
			continue
		}

		if region != "EN" {
			url := fmt.Sprintf("https://art.gametdb.com/ds/cover%s/EN/%s%s",
				q.quality, titleID, q.ext)
			data, err := Download(url)
			if err == nil {
				return data, nil
			}
		}
	}

	return nil, ErrNotFound
}
