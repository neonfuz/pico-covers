package database

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

var NoIntroConsoleStrings = map[ConsoleType]string{
	NES:                    "Nintendo - Nintendo Entertainment System",
	SNES:                   "Nintendo - Super Nintendo Entertainment System",
	GB:                     "Nintendo - Game Boy",
	GBC:                    "Nintendo - Game Boy Color",
	GBA:                    "Nintendo - Game Boy Advance",
	NDS:                    "Nintendo - Nintendo DS",
	DSi:                    "Nintendo - Nintendo DSi",
	GameGear:               "Sega - Game Gear",
	Genesis:                "Sega - Mega Drive - Genesis",
	MasterSystem:           "Sega - Master System - Mark III",
	FDS:                    "Nintendo - Famicom Disk System",
	NintendoDSDownloadPlay: "Nintendo - Nintendo DS (Download Play)",
	NintendoDSiDigital:     "Nintendo - Nintendo DSi (Digital)",
}

const noIntroURL = "https://datomatic.no-intro.org/index.php?page=download&fun=wut"

type dataFileXML struct {
	XMLName xml.Name  `xml:"datafile"`
	Games   []gameXML `xml:"game"`
}

type gameXML struct {
	XMLName xml.Name `xml:"game"`
	Name    string   `xml:"name,attr"`
	GameId  string   `xml:"game_id,attr"`
	Rom     romXML   `xml:"rom"`
}

type romXML struct {
	Sha1   string `xml:"sha1,attr"`
	Serial string `xml:"serial,attr"`
	Status string `xml:"status,attr"`
}

func DownloadNoIntro(ctx context.Context, consoleType ConsoleType) ([]RomMetaData, error) {
	consoleStr, ok := NoIntroConsoleStrings[consoleType]
	if !ok {
		return nil, fmt.Errorf("no NoIntro console string for type %v", consoleType)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Jar:       jar,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, noIntroURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating GET request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request: %w", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	form := url.Values{}
	form.Set("download", "Download")
	form.Set("sel_s", consoleStr)

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, noIntroURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("reading zip response: %w", err)
	}

	if len(zipReader.File) == 0 {
		return nil, fmt.Errorf("no files in downloaded zip")
	}

	xmlFile, err := zipReader.File[0].Open()
	if err != nil {
		return nil, fmt.Errorf("opening xml in zip: %w", err)
	}
	defer xmlFile.Close()

	xmlData, err := io.ReadAll(xmlFile)
	if err != nil {
		return nil, fmt.Errorf("reading xml from zip: %w", err)
	}

	return parseNoIntroXML(xmlData, consoleType)
}

func parseNoIntroXML(data []byte, consoleType ConsoleType) ([]RomMetaData, error) {
	var df dataFileXML
	if err := xml.Unmarshal(data, &df); err != nil {
		return nil, fmt.Errorf("parsing XML: %w", err)
	}

	baseType := consoleType.BaseType()
	var results []RomMetaData

	for _, game := range df.Games {
		results = append(results, RomMetaData{
			ConsoleType:    baseType,
			ConsoleSubType: consoleType,
			Name:           game.Name,
			Sha1:           strings.ToLower(game.Rom.Sha1),
			Serial:         game.Rom.Serial,
			GameId:         game.GameId,
			Status:         game.Rom.Status,
		})
	}

	return results, nil
}
