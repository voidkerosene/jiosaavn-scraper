package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// routes registers every endpoint. All service handlers match the same
// method signature, so wiring is a one-liner per route.
func routes(c *cache) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", direct(func(w http.ResponseWriter, r *http.Request) {
		w.Write(render(200, getDocs()))
	}))

	mux.HandleFunc("GET /api/search", pipeline(searchAll, c))
	mux.HandleFunc("GET /api/search/songs", pipeline(searchSongs, c))
	mux.HandleFunc("GET /api/search/albums", pipeline(searchAlbums, c))
	mux.HandleFunc("GET /api/search/artists", pipeline(searchArtists, c))
	mux.HandleFunc("GET /api/search/playlists", pipeline(searchPlaylists, c))
	mux.HandleFunc("GET /api/songs", pipeline(songs, c))
	mux.HandleFunc("GET /api/songs/{id}", pipeline(songByID, c))
	mux.HandleFunc("GET /api/songs/{id}/suggestions", pipeline(songSuggestions, c))
	mux.HandleFunc("GET /api/albums", pipeline(albums, c))
	mux.HandleFunc("GET /api/playlists", pipeline(playlists, c))
	mux.HandleFunc("GET /api/artists", pipeline(artists, c))
	mux.HandleFunc("GET /api/artists/{id}", pipeline(artistByID, c))
	mux.HandleFunc("GET /api/artists/{id}/songs", pipeline(artistSongs, c))
	mux.HandleFunc("GET /api/artists/{id}/albums", pipeline(artistAlbums, c))
	mux.HandleFunc("GET /api/trending", pipeline(trending, c))

	// Discover: the JioSaavn home page, language-scoped.
	mux.HandleFunc("GET /api/browse", pipeline(browse, c))
	mux.HandleFunc("GET /api/stations", pipeline(browseSection("stations"), c))
	mux.HandleFunc("GET /api/channels", pipeline(browseSection("channels"), c))
	mux.HandleFunc("GET /api/charts", pipeline(browseSection("charts"), c))
	mux.HandleFunc("GET /api/new-releases", pipeline(browseSection("newAlbums"), c))
	mux.HandleFunc("GET /api/new-trending", pipeline(browseSection("newTrending"), c))
	mux.HandleFunc("GET /api/top-playlists", pipeline(topPlaylists, c))
	mux.HandleFunc("GET /api/top-shows", pipeline(browseSection("topShows"), c))

	// Lyrics via LRCLib.
	mux.HandleFunc("GET /api/lyrics", pipeline(lyricsQuery, c))
	mux.HandleFunc("GET /api/songs/{id}/lyrics", pipeline(songLyrics, c))

	// JSON 404 for anything else.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusNotFound)
		message := "route not found, all endpoints are documented at GET /"
		if _, err := url.Parse(r.URL.Path); err == nil && r.URL.Path == "/api" {
			message = "api root has no content, all endpoints are documented at GET /"
		}
		w.Write(render(http.StatusNotFound, errorPayload{Success: false, Message: message}))
	})

	return mux
}

// Serve builds the full stack and blocks on ListenAndServe.
func Serve(cfg Config) error {
	c := newCache(cfg.CacheTTL, cfg.CacheMaxSize)

	// Log sink: stdout always, plus LOG_FILE when configured.
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("log file %s: %w", cfg.LogFile, err)
		}
		log.SetOutput(io.MultiWriter(os.Stderr, f))
		defer f.Close()
	}

	addr := cfg.Addr
	log.Printf("JioSaavn Scrape API listening on http://localhost%s (docs at /)", addr)
	return http.ListenAndServe(addr, routes(c))
}
