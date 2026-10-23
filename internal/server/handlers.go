package server

import (
	"net/http"
	"strconv"
	"strings"

	"jiosaavn-scraper/internal/api"
	"jiosaavn-scraper/internal/lrclib"
)

// Each handler pulls its params off the request and delegates to the
// api service. No ResponseWriter access — the pipeline handles the rest.

func q(req *http.Request, key string) string { return req.URL.Query().Get(key) }

// positive: query params like limit/page must not crash on garbage
// input (the original used naive Number() coercion and let NaN through).
func positive(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return def
	}
	return n
}

func searchAll(r *http.Request) (any, error) {
	return api.SearchAll(r.Context(), q(r, "query"))
}

func searchSongs(r *http.Request) (any, error) {
	return api.SearchSongs(r.Context(), q(r, "query"), positive(q(r, "page"), 0), positive(q(r, "limit"), 10))
}

func searchAlbums(r *http.Request) (any, error) {
	return api.SearchAlbums(r.Context(), q(r, "query"), positive(q(r, "page"), 0), positive(q(r, "limit"), 10))
}

func searchArtists(r *http.Request) (any, error) {
	return api.SearchArtists(r.Context(), q(r, "query"), positive(q(r, "page"), 0), positive(q(r, "limit"), 10))
}

func searchPlaylists(r *http.Request) (any, error) {
	return api.SearchPlaylists(r.Context(), q(r, "query"), positive(q(r, "page"), 0), positive(q(r, "limit"), 10))
}

func songs(r *http.Request) (any, error) {
	return api.GetSong(r.Context(), q(r, "ids"), q(r, "link"))
}

func songByID(r *http.Request) (any, error) {
	return api.GetSong(r.Context(), r.PathValue("id"), "")
}

func songSuggestions(r *http.Request) (any, error) {
	return api.GetSongSuggestions(r.Context(), r.PathValue("id"), positive(q(r, "limit"), 10))
}

func albums(r *http.Request) (any, error) {
	return api.GetAlbum(r.Context(), q(r, "id"), q(r, "link"))
}

func playlists(r *http.Request) (any, error) {
	return api.GetPlaylist(r.Context(), q(r, "id"), q(r, "link"), positive(q(r, "page"), 0), positive(q(r, "limit"), 10))
}

func artists(r *http.Request) (any, error) {
	return api.GetArtist(r.Context(), q(r, "id"), q(r, "link"), positive(q(r, "page"), 0), positive(q(r, "songCount"), 10), positive(q(r, "albumCount"), 10), q(r, "sortBy"), q(r, "sortOrder"))
}

func artistByID(r *http.Request) (any, error) {
	return api.GetArtist(r.Context(), r.PathValue("id"), "", positive(q(r, "page"), 0), positive(q(r, "songCount"), 10), positive(q(r, "albumCount"), 10), q(r, "sortBy"), q(r, "sortOrder"))
}

func artistSongs(r *http.Request) (any, error) {
	return api.GetArtistSongs(r.Context(), r.PathValue("id"), positive(q(r, "page"), 0), q(r, "sortBy"), q(r, "sortOrder"))
}

func artistAlbums(r *http.Request) (any, error) {
	return api.GetArtistAlbums(r.Context(), r.PathValue("id"), positive(q(r, "page"), 0), q(r, "sortBy"), q(r, "sortOrder"))
}

func trending(r *http.Request) (any, error) {
	return api.Trending(r.Context(), q(r, "type"), q(r, "language"))
}

// ---------- discover (browse) endpoints ----------

// browseSection pins a handler to one browse section; it lets the
// friendly aliases (/api/charts, /api/stations, ...) share one
// implementation with /api/browse?section=...
func browseSection(section string) kinded {
	return func(r *http.Request) (any, error) {
		return api.Browse(r.Context(), q(r, "language"), section)
	}
}

func browse(r *http.Request) (any, error) {
	return api.Browse(r.Context(), q(r, "language"), q(r, "section"))
}

func topPlaylists(r *http.Request) (any, error) {
	return api.FeaturedPlaylists(r.Context(), q(r, "language"), positive(q(r, "page"), 0), positive(q(r, "limit"), 10))
}

// ---------- lyrics (LRCLib integration) ----------

type lyricsPayload struct {
	SongID       string  `json:"songId"`
	Song         string  `json:"song"`
	Artist       *string `json:"artist"`
	Album        *string `json:"album"`
	Duration     int     `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	SyncedLyrics *string `json:"syncedLyrics"`
	PlainLyrics  *string `json:"plainLyrics"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    *string `json:"albumName"`
	LrclibID     int     `json:"lrclibId"`
}

func songLyrics(r *http.Request) (any, error) {
	songs, err := api.GetSong(r.Context(), r.PathValue("id"), "")
	if err != nil {
		return nil, err
	}
	if len(songs) == 0 {
		return nil, &api.HTTPError{Code: 404, Message: "song not found"}
	}
	s := songs[0]

	artist := ""
	if len(s.Artists.Primary) > 0 {
		artist = s.Artists.Primary[0].Name
	} else if len(s.Artists.All) > 0 {
		artist = s.Artists.All[0].Name
	}
	album := ""
	if s.Album.Name != nil {
		album = *s.Album.Name
	}
	duration := 0
	if s.Duration != nil {
		duration = *s.Duration
	}

	lyr, lerr := lrclib.Get(r.Context(), s.Name, artist, album, duration)
	if lerr != nil && !lrclib.IsNotFound(lerr) {
		return nil, lerr
	}
	if lyr == nil {
		return nil, &api.HTTPError{Code: 404, Message: "no lyrics found for this song on LRCLib"}
	}

	var artistPtr *string
	if artist != "" {
		artistPtr = &artist
	}
	return lyricsPayload{
		SongID:       s.ID,
		Song:         s.Name,
		Artist:       artistPtr,
		Album:        s.Album.Name,
		Duration:     duration,
		Instrumental: lyr.Instrumental,
		SyncedLyrics: lyr.SyncedLyrics,
		PlainLyrics:  lyr.PlainLyrics,
		TrackName:    lyr.TrackName,
		ArtistName:   lyr.ArtistName,
		AlbumName:    lyr.AlbumName,
		LrclibID:     lyr.ID,
	}, nil
}

func lyricsQuery(r *http.Request) (any, error) {
	query := strings.TrimSpace(q(r, "query"))
	if query == "" {
		return nil, &api.HTTPError{Code: 400, Message: "search query is required"}
	}
	results, err := lrclib.Search(r.Context(), query)
	if err != nil && !lrclib.IsNotFound(err) {
		return nil, err
	}
	if results == nil {
		results = []lrclib.Lyrics{}
	}
	return results, nil
}
