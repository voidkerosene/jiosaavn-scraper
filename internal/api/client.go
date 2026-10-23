package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"
)

const upstreamBase = "https://www.jiosaavn.com/api.php"

// Upstream endpoint calls, identical to the original API.
const (
	callSearchAll         = "autocomplete.get"
	callSearchSongs       = "search.getResults"
	callSearchAlbums      = "search.getAlbumResults"
	callSearchArtists     = "search.getArtistResults"
	callSearchPlaylists   = "search.getPlaylistResults"
	callSongDetails       = "song.getDetails"
	callWebAPI            = "webapi.get"
	callSongSuggestions   = "webradio.getSong"
	callCreateStation     = "webradio.createEntityStation"
	callAlbumDetails      = "content.getAlbumDetails"
	callPlaylistDetails   = "playlist.getDetails"
	callArtistDetails     = "artist.getArtistPageDetails"
	callArtistSongs       = "artist.getArtistMoreSong"
	callArtistAlbums      = "artist.getArtistMoreAlbum"
	callTrending          = "content.getTrending"
	callBrowseModules     = "content.getBrowseModules"
	callFeaturedPlaylists = "playlist.getFeaturedPlaylists"
)

// userAgents is a small rotation of common desktop agents. JioSaavn
// occasionally refuses the default Go UA.
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0",
}

// shared client: keep-alive connections pooled per host so repeat
// requests skip TCP + TLS handshakes entirely.
var httpClient = &http.Client{
	Timeout: 20 * time.Second,
	Transport: &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 8 * time.Second}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

// Fetch calls an upstream __call endpoint with e=xtra params and decodes
// the JSON response into out.
func fetch(ctx context.Context, call string, params map[string]string, out any) error {
	q := url.Values{}
	q.Set("__call", call)
	q.Set("_format", "json")
	q.Set("_marker", "0")
	q.Set("api_version", "4")
	q.Set("ctx", "web6dot0")
	for k, v := range params {
		q.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamBase+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upstream %s: %w", call, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream %s: status %d", call, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return fmt.Errorf("upstream %s: %w", call, err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("upstream %s: bad JSON: %w", call, err)
	}
	return nil
}
