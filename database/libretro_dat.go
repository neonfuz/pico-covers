package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var LibRetroDatURLs = map[ConsoleType]string{
	NES:          "https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Nintendo%20Entertainment%20System.dat",
	SNES:         "https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Super%20Nintendo%20Entertainment%20System.dat",
	GB:           "https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Game%20Boy.dat",
	GBC:          "https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Game%20Boy%20Color.dat",
	GBA:          "https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Game%20Boy%20Advance.dat",
	NDS:          "https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Nintendo%20DS.dat",
	DSi:          "https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Nintendo%20DSi.dat",
	GameGear:     "https://github.com/libretro/libretro-database/raw/master/dat/Sega%20-%20Game%20Gear.dat",
	Genesis:      "https://github.com/libretro/libretro-database/raw/master/dat/Sega%20-%20Mega%20Drive%20-%20Genesis.dat",
	MasterSystem: "https://github.com/libretro/libretro-database/raw/master/dat/Sega%20-%20Master%20System%20-%20Mark%20III.dat",
	FDS:          "https://github.com/libretro/libretro-database/raw/master/dat/Nintendo%20-%20Famicom%20Disk%20System.dat",
}

var libRetroClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func ParseLibRetroDat(ctx context.Context, consoleType ConsoleType) ([]RomMetaData, error) {
	baseType := consoleType.BaseType()
	url, ok := LibRetroDatURLs[baseType]
	if !ok {
		return nil, fmt.Errorf("no LibRetro DAT URL for console type %v", consoleType)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := libRetroClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching LibRetro DAT: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return parseSExpr(string(body), consoleType)
}

func tokenize(input string) []string {
	var tokens []string
	i := 0
	for i < len(input) {
		switch {
		case input[i] == ' ' || input[i] == '\t' || input[i] == '\n' || input[i] == '\r':
			i++
		case input[i] == '(' || input[i] == ')':
			tokens = append(tokens, string(input[i]))
			i++
		case input[i] == '"':
			i++
			start := i
			for i < len(input) && input[i] != '"' {
				i++
			}
			tokens = append(tokens, input[start:i])
			if i < len(input) {
				i++
			}
		default:
			start := i
			for i < len(input) && input[i] != ' ' && input[i] != '\t' && input[i] != '\n' && input[i] != '\r' && input[i] != '(' && input[i] != ')' && input[i] != '"' {
				i++
			}
			tokens = append(tokens, input[start:i])
		}
	}
	return tokens
}

type tokenReader struct {
	tokens []string
	pos    int
}

func (r *tokenReader) peek() string {
	if r.pos < len(r.tokens) {
		return r.tokens[r.pos]
	}
	return ""
}

func (r *tokenReader) next() string {
	t := r.peek()
	r.pos++
	return t
}

func (r *tokenReader) expect(s string) error {
	if r.next() != s {
		return fmt.Errorf("expected %q", s)
	}
	return nil
}

func (r *tokenReader) done() bool {
	return r.pos >= len(r.tokens)
}

func parseSExpr(input string, consoleType ConsoleType) ([]RomMetaData, error) {
	r := &tokenReader{tokens: tokenize(input)}
	baseType := consoleType.BaseType()
	var results []RomMetaData

	for !r.done() {
		if r.peek() == "game" {
			r.next()
			r.expect("(")
			meta := RomMetaData{
				ConsoleType:    baseType,
				ConsoleSubType: consoleType,
			}
			for r.peek() != ")" && !r.done() {
				key := r.next()
				switch key {
				case "name":
					meta.Name = r.next()
				case "rom":
					r.expect("(")
					for r.peek() != ")" && !r.done() {
						rkey := r.next()
						switch rkey {
						case "sha1":
							meta.Sha1 = strings.ToLower(r.next())
						default:
							r.next()
						}
					}
					r.expect(")")
				default:
					r.next()
				}
			}
			r.expect(")")
			if meta.Name != "" {
				results = append(results, meta)
			}
		} else {
			r.next()
		}
	}
	return results, nil
}
