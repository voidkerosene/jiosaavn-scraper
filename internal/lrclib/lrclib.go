// Package lrclib talks to the free LRCLib lyrics API
// (https://lrclib.net no key required, fair-use rate limits).
// User-Agent etiquette per their docs, and rely on server-side caching
// to stay friendly to the public API.
package lrclib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const base = "https://lrclib.net"

// Lyrics is one LRCLib match. LRCLib sends duration as a JSON float
// ("204.0"), which fails to decode into int.
type Lyrics struct {
	ID            int     `json:"id"`
	TrackName     string  `json:"trackName"`
	ArtistName    string  `json:"artistName"`
	AlbumName     *string `json:"albumName"`
	DurationFloat float64 `json:"duration"`
	Instrumental  bool    `json:"instrumental"`
	PlainLyrics   *string `json:"plainLyrics"`
	SyncedLyrics  *string `json:"syncedLyrics"`
}

// Duration returns the track length as whole seconds.
func (l *Lyrics) Duration() int {
	return int(l.DurationFloat + 0.5)
}

// errNotFound is the LRCLib-miss sentinel.
var errNotFound = fmt.Errorf("lrclib: not found")

// IsNotFound reports whether err came from LRCLib's 404.
func IsNotFound(err error) bool { return err == errNotFound }

var client = &http.Client{Timeout: 15 * time.Second}

func get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "jiosaavn-scraper/1.0 (music metadata proxy; LRCLib client)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lrclib %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	}
	// ponytail: 429 is deliberately NOT shared with waiters as a
	// cacheable failure; a plain error keeps it outside the 5-min cache
	// so a hot 429 storm cannot get stuck in cache.
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("lrclib: rate limited, try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lrclib %s: status %d", path, resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}

// Get returns the best match for a song: /api/get first (exact), then a
// ranked /api/search fallback picking the closest duration, since
// JioSaavn titles like `Song (From "Movie")` rarely exact-match.
func Get(ctx context.Context, trackName, artistName, albumName string, duration int) (*Lyrics, error) {
	p := url.Values{"track_name": {trackName}}
	if artistName != "" {
		p.Set("artist_name", artistName)
	}
	if albumName != "" {
		p.Set("album_name", albumName)
	}
	if duration > 0 {
		p.Set("duration", strconv.Itoa(duration))
	}

	if body, err := get(ctx, "/api/get", p); err == nil {
		var l Lyrics
		if json.Unmarshal(body, &l) == nil && l.ID != 0 {
			return &l, nil
		}
	} else if !IsNotFound(err) {
		return nil, err
	}

	results, err := Search(ctx, trackName+" "+artistName)
	if err != nil {
		return nil, err
	}
	return bestMatch(results, duration), nil
}

// bestMatch picks the result whose duration is closest to the song's;
// anything more than ~15s off is treated as a different recording.
func bestMatch(results []Lyrics, duration int) *Lyrics {
	if len(results) == 0 {
		return nil
	}
	if duration <= 0 {
		return &results[0] // first-ranked result when duration unknown
	}
	best, bestDiff := results[0], absInt(duration-results[0].Duration())
	for _, r := range results[1:] {
		if d := absInt(duration - r.Duration()); d < bestDiff {
			best, bestDiff = r, d
		}
	}
	if bestDiff > 15 {
		return nil
	}
	return &best
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Search is the raw LRCLib search, used by /api/lyrics?query=.
func Search(ctx context.Context, query string) ([]Lyrics, error) {
	body, err := get(ctx, "/api/search", url.Values{"q": {query}})
	if err != nil {
		return nil, err
	}
	var list []Lyrics
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("lrclib search: bad JSON: %w", err)
	}
	return list, nil
}
