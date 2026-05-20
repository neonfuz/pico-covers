package database

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/neonfuz/pico-covers/events"
)

type Database struct {
	Records   []RomMetaData
	sha1Index map[string][]int
	mu        sync.RWMutex
}

const defaultCacheSubPath = ".cache/pico-covers/nointro.db"

func NewDatabase() *Database {
	return &Database{
		sha1Index: make(map[string][]int),
	}
}

func (db *Database) Initialize(ctx context.Context, cachePath string, forceRefresh bool, handler events.EventHandler) error {
	if cachePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		cachePath = filepath.Join(home, defaultCacheSubPath)
	}

	if !forceRefresh {
		if err := db.loadCache(cachePath); err == nil {
			if handler != nil {
				handler(events.ProgressEvent{Kind: events.EventDBLoaded, Total: len(db.Records)})
			}
			return nil
		}
	}

	if handler != nil {
		handler(events.ProgressEvent{Kind: events.EventDBInit, Detail: "Downloading NoIntro database..."})
	}

	var allRecords []RomMetaData

	for ct := range NoIntroConsoleStrings {
		records, err := DownloadNoIntro(ctx, ct)
		if err != nil {
			continue
		}
		allRecords = append(allRecords, records...)
	}

	for ct := range LibRetroDatURLs {
		records, err := ParseLibRetroDat(ctx, ct)
		if err != nil {
			continue
		}
		allRecords = append(allRecords, records...)
	}

	db.mu.Lock()
	db.Records = allRecords
	db.buildIndex()
	db.mu.Unlock()

	if err := db.saveCache(cachePath); err != nil {
		return fmt.Errorf("saving cache: %w", err)
	}

	if handler != nil {
		handler(events.ProgressEvent{Kind: events.EventDBLoaded, Total: len(db.Records)})
	}

	return nil
}

func (db *Database) buildIndex() {
	db.sha1Index = make(map[string][]int, len(db.Records)/2)
	for i, rec := range db.Records {
		sha := strings.ToLower(rec.Sha1)
		if sha == "" {
			continue
		}
		db.sha1Index[sha] = append(db.sha1Index[sha], i)
	}
}

func (db *Database) loadCache(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()

	data, err := io.ReadAll(gz)
	if err != nil {
		return err
	}

	var records []RomMetaData
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	db.mu.Lock()
	db.Records = records
	db.buildIndex()
	db.mu.Unlock()

	return nil
}

func (db *Database) saveCache(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gz := gzip.NewWriter(f)
	defer func() { _ = gz.Close() }()

	data, err := json.Marshal(db.Records)
	if err != nil {
		return err
	}

	if _, err := gz.Write(data); err != nil {
		return err
	}

	return gz.Close()
}

func (db *Database) FindBySHA1(sha1 string) *RomMetaData {
	db.mu.RLock()
	defer db.mu.RUnlock()

	sha := strings.ToLower(strings.TrimSpace(sha1))
	indices, ok := db.sha1Index[sha]
	if !ok || len(indices) == 0 {
		return nil
	}
	return &db.Records[indices[0]]
}

func (db *Database) FindByTitleID(titleID string, consoleType ConsoleType) *RomMetaData {
	db.mu.RLock()
	defer db.mu.RUnlock()

	titleID = strings.TrimSpace(titleID)
	if titleID == "" {
		return nil
	}

	baseType := consoleType.BaseType()

	var badMatch *RomMetaData
	for i := range db.Records {
		rec := &db.Records[i]
		if rec.ConsoleType != baseType {
			continue
		}
		if !strings.EqualFold(rec.Serial, titleID) && !strings.EqualFold(rec.GameId, titleID) {
			continue
		}
		if strings.EqualFold(rec.Status, "bad") {
			if badMatch == nil {
				badMatch = rec
			}
			continue
		}
		return rec
	}

	return badMatch
}
