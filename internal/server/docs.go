package server

// apiDocs are grouped the way a streaming-site developer wires an app:
// discovery feeds first, then entity search, entity details, and
// cross-cutting services (lyrics).

type apiDoc struct {
	Group       string        `json:"group"`
	Path        string        `json:"path"`
	Method      string        `json:"method"`
	Summary     string        `json:"summary"`
	Description string        `json:"description,omitempty"`
	Params      []apiDocParam `json:"params"`
}

type apiDocParam struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

func param(name, typ, desc string, required bool, def ...string) apiDocParam {
	p := apiDocParam{Name: name, Required: required, Type: typ, Description: desc}
	if len(def) > 0 {
		p.Default = def[0]
	}
	return p
}

const languageParamDesc = "Content language, e.g. hindi, english, punjabi, tamil, telugu, malayalam, kannada, bengali, marathi, urdu, bhojpuri, haryanvi, rajasthani, korean, arabic"

func languageParams() []apiDocParam {
	return []apiDocParam{param("language", "string", languageParamDesc, false, "hindi")}
}

var docs = []apiDoc{
	// Discover: feeds a streaming-site home page.
	{
		Group: "Discover", Path: "/api/browse", Method: "GET",
		Summary:     "Full JioSaavn home page for a language",
		Description: "Everything jiosaavn.com shows: radio stations, mood/genre channels, charts, new albums, new trending, top playlists, top shows. Use section=... to request one section only for a lighter payload.",
		Params: append([]apiDocParam{
			param("section", "string", "all | stations | channels | charts | newAlbums | newTrending | topPlaylists | topShows", false, "all"),
		}, languageParams()...),
	},
	{
		Group: "Discover", Path: "/api/stations", Method: "GET",
		Summary: "Featured radio stations for a language",
		Params:  languageParams(),
	},
	{
		Group: "Discover", Path: "/api/channels", Method: "GET",
		Summary:     "Mood/genre discover channels (e.g. Romance, Workout)",
		Description: "Genre and mood classifications straight from JioSaavn's editorial taxonomy, with subType on each channel.",
		Params:      languageParams(),
	},
	{
		Group: "Discover", Path: "/api/charts", Method: "GET",
		Summary:     "Official charts for a language",
		Description: "Chart playlists (e.g. India Superhits Top 50). Feed their id into /api/playlists?id=<chartId> to get the ranked songs.",
		Params:      languageParams(),
	},
	{
		Group: "Discover", Path: "/api/new-releases", Method: "GET",
		Summary: "New album releases for a language",
		Params:  languageParams(),
	},
	{
		Group: "Discover", Path: "/api/new-trending", Method: "GET",
		Summary: "Freshly trending playlists for a language",
		Params:  languageParams(),
	},
	{
		Group: "Discover", Path: "/api/top-playlists", Method: "GET",
		Summary:     "Top/featured playlists for a language",
		Description: "Language-aware curated feed (dedicated upstream call), with follower counts and types — the per-language rail behind the JioSaavn app's home page.",
		Params: append([]apiDocParam{
			param("page", "number", "Page number (0-based)", false, "0"),
			param("limit", "number", "Results per page (max 250)", false, "10"),
		}, languageParams()...),
	},
	{
		Group: "Discover", Path: "/api/top-shows", Method: "GET",
		Summary: "Top shows/podcasts for a language",
		Params:  languageParams(),
	},
	{
		Group: "Discover", Path: "/api/trending", Method: "GET",
		Summary:     "Trending entities (the top-hits feed)",
		Description: "Type- and language-aware: type maps to the upstream entity_type and language to entity_language, cached per query so combinations never share a response.",
		Params: append([]apiDocParam{
			param("type", "string", "album | song | playlist", false, "album"),
		}, languageParams()...),
	},

	// Search.
	{
		Group: "Search", Path: "/api/search", Method: "GET",
		Summary:     "Global search",
		Description: "Search songs, albums, artists, and playlists in one call.",
		Params:      []apiDocParam{param("query", "string", "Search query", true)},
	},
	{
		Group: "Search", Path: "/api/search/songs", Method: "GET",
		Summary: "Search for songs",
		Params:  searchParams(),
	},
	{
		Group: "Search", Path: "/api/search/albums", Method: "GET",
		Summary: "Search for albums",
		Params:  searchParams(),
	},
	{
		Group: "Search", Path: "/api/search/artists", Method: "GET",
		Summary: "Search for artists",
		Params:  searchParams(),
	},
	{
		Group: "Search", Path: "/api/search/playlists", Method: "GET",
		Summary: "Search for playlists",
		Params:  searchParams(),
	},

	// Songs.
	{
		Group: "Songs", Path: "/api/songs", Method: "GET",
		Summary:     "Retrieve songs by IDs or link",
		Description: "Comma-separated song IDs, or a song permalink. downloadUrl contains one entry per bitrate present for the master (320kbps only when the upstream flag says the master has it).",
		Params: []apiDocParam{
			param("ids", "string", "Comma-separated song IDs", false),
			param("link", "string", "JioSaavn song link", false),
		},
	},
	{
		Group: "Songs", Path: "/api/songs/{id}", Method: "GET",
		Summary: "Retrieve a song by ID",
		Params:  []apiDocParam{},
	},
	{
		Group: "Songs", Path: "/api/songs/{id}/suggestions", Method: "GET",
		Summary:     "Song suggestions for infinite playback",
		Description: "Creates a radio station for the song server-side and returns similar songs. Upstream radio quota is flaky in some regions; on failure a clean 404 is returned.",
		Params: []apiDocParam{
			param("limit", "number", "Number of suggestions (max 100)", false, "10"),
		},
	},

	// Albums / playlists / artists.
	{
		Group: "Albums", Path: "/api/albums", Method: "GET",
		Summary: "Retrieve an album by ID or link",
		Params: []apiDocParam{
			param("id", "string", "The album ID", false),
			param("link", "string", "JioSaavn album link", false),
		},
	},
	{
		Group: "Playlists", Path: "/api/playlists", Method: "GET",
		Summary:     "Retrieve a playlist by ID or link",
		Description: "Also accepts chart IDs from /api/charts to fetch ranked chart songs.",
		Params: []apiDocParam{
			param("id", "string", "The playlist ID", false),
			param("link", "string", "JioSaavn playlist link", false),
			param("page", "number", "Page of songs to retrieve", false, "0"),
			param("limit", "number", "Number of songs per page", false, "10"),
		},
	},
	{
		Group: "Artists", Path: "/api/artists", Method: "GET",
		Summary: "Retrieve an artist by ID or link",
		Params: []apiDocParam{
			param("id", "string", "The artist ID", false),
			param("link", "string", "JioSaavn artist link", false),
			param("page", "number", "Page number", false, "0"),
			param("songCount", "number", "Number of top songs", false, "10"),
			param("albumCount", "number", "Number of top albums", false, "10"),
			param("sortBy", "string", "popularity | latest | alphabetical", false, "popularity"),
			param("sortOrder", "string", "asc | desc", false, "asc"),
		},
	},
	{
		Group: "Artists", Path: "/api/artists/{id}", Method: "GET",
		Summary: "Retrieve an artist by ID",
		Params: []apiDocParam{
			param("songCount", "number", "Number of top songs", false, "10"),
			param("albumCount", "number", "Number of top albums", false, "10"),
		},
	},
	{
		Group: "Artists", Path: "/api/artists/{id}/songs", Method: "GET",
		Summary: "An artist's songs",
		Params: []apiDocParam{
			param("sortBy", "string", "popularity | latest | alphabetical", false, "popularity"),
			param("sortOrder", "string", "asc | desc", false, "desc"),
		},
	},
	{
		Group: "Artists", Path: "/api/artists/{id}/albums", Method: "GET",
		Summary: "An artist's albums",
		Params: []apiDocParam{
			param("sortBy", "string", "popularity | latest | alphabetical", false, "popularity"),
			param("sortOrder", "string", "asc | desc", false, "desc"),
		},
	},

	// Lyrics (LRCLib).
	{
		Group: "Lyrics", Path: "/api/songs/{id}/lyrics", Method: "GET",
		Summary:     "Synced + plain lyrics for a song",
		Description: "Looks the song up on JioSaavn, then matches it on LRCLib (free lyrics API) using track/artist/album/duration, with a duration-aware fallback search.",
		Params:      []apiDocParam{},
	},
	{
		Group: "Lyrics", Path: "/api/lyrics", Method: "GET",
		Summary:     "Direct LRCLib lyrics search",
		Description: "Raw LRCLib matches with syncedLyrics (LRC format) and plainLyrics fields; useful when you already know the track/artist.",
		Params: []apiDocParam{
			param("query", "string", "Free-form lyrics search, e.g. 'Como La Flor' or 'artist track'", true),
		},
	},
}

func searchParams() []apiDocParam {
	return []apiDocParam{
		param("query", "string", "Search query", true),
		param("page", "number", "Result page number (0-based)", false, "0"),
		param("limit", "number", "Results per page (max 100)", false, "10"),
	}
}

type docsResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Language    string   `json:"language"`
	Upstream    string   `json:"upstream"`
	Notes       []string `json:"notes"`
	Endpoints   []apiDoc `json:"endpoints"`
}

func getDocs() docsResponse {
	return docsResponse{
		Name:        "JioSaavn Scrape API",
		Description: "Unofficial API for JioSaavn: search, browse, trending, charts, stations, lyrics, and song metadata with audio stream URLs at every available quality. Every response is JSON in a {success, data} envelope. Endpoint groups map to the sections of a music streaming site.",
		Version:     "1.1.0",
		Language:    "Go",
		Upstream:    "https://www.jiosaavn.com/api.php (+ https://lrclib.net for lyrics)",
		Notes: []string{
			"All endpoints live under /api and are GET-only.",
			"Responses are cached for 5 minutes per full query string and gzip-compressed when requested; concurrent identical requests share one upstream fetch.",
			"Errors return {\"success\": false, \"message\": \"...\"} with a matching HTTP status.",
			"X-Cache header: hit / miss / bypass (errors are never cached).",
			"downloadUrl arrays contain one entry per available bitrate (12/48/96/160/320 kbps).",
		},
		Endpoints: docs,
	}
}
