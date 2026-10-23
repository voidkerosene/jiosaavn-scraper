package api

// Browse sections mirror the JioSaavn home page: one upstream call
// (content.getBrowseModules) returns everything, language-scoped.

type BrowseSection string

const (
	SectionAll          BrowseSection = "all"
	SectionStations     BrowseSection = "stations"
	SectionChannels     BrowseSection = "channels"
	SectionCharts       BrowseSection = "charts"
	SectionNewAlbums    BrowseSection = "newAlbums"
	SectionNewTrending  BrowseSection = "newTrending"
	SectionTopPlaylists BrowseSection = "topPlaylists"
	SectionTopShows     BrowseSection = "topShows"
)

func sectionAllowed(s BrowseSection) bool {
	switch s {
	case SectionAll, SectionStations, SectionChannels, SectionCharts,
		SectionNewAlbums, SectionNewTrending, SectionTopPlaylists, SectionTopShows:
		return true
	}
	return false
}

// BrowseStationRaw etc. keep only the fields each section actually has.
type BrowseStationRaw struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Subtitle        string `json:"subtitle"`
	Image           string `json:"image"`
	PermaURL        string `json:"perma_url"`
	Language        string `json:"language"`
	ExplicitContent string `json:"explicit_content"`
	MoreInfo        struct {
		SubType  string `json:"sub_type"`
		Language string `json:"language"`
	} `json:"more_info"`
}

type BrowseChannelRaw struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Subtitle        string `json:"subtitle"`
	Type            string `json:"type"`
	Image           string `json:"image"`
	PermaURL        string `json:"perma_url"`
	ExplicitContent string `json:"explicit_content"`
	MoreInfo        struct {
		SubType string `json:"sub_type"`
	} `json:"more_info"`
}

type BrowseChartRaw struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Subtitle        string `json:"subtitle"`
	Image           string `json:"image"`
	PermaURL        string `json:"perma_url"`
	Language        string `json:"language"`
	ExplicitContent string `json:"explicit_content"`
}

type BrowsePlaylistRaw struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Subtitle        string         `json:"subtitle"`
	HeaderDesc      string         `json:"header_desc"`
	Image           string         `json:"image"`
	PermaURL        string         `json:"perma_url"`
	Language        string         `json:"language"`
	Year            StringOrNumber `json:"year"`
	ExplicitContent string         `json:"explicit_content"`
	ListCount       StringOrNumber `json:"list_count"`
	MoreInfo        struct {
		SongCount     StringOrNumber `json:"song_count"`
		FollowerCount StringOrNumber `json:"follower_count"`
		FirstName     string         `json:"firstname"`
	} `json:"more_info"`
}

type BrowseShowRaw struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Image    string `json:"image"`
	PermaURL string `json:"perma_url"`
	Language string `json:"language"`
}

type browseModulesRaw struct {
	Radio struct {
		FeaturedStations []BrowseStationRaw `json:"featured_stations"`
	} `json:"radio"`
	BrowseDiscover []BrowseChannelRaw  `json:"browse_discover"`
	Charts         []BrowseChartRaw    `json:"charts"`
	NewAlbums      []AlbumRaw          `json:"new_albums"`
	NewTrending    []BrowsePlaylistRaw `json:"new_trending"`
	TopPlaylists   []BrowsePlaylistRaw `json:"top_playlists"`
	TopShows       struct {
		Shows []BrowseShowRaw `json:"shows"`
	} `json:"top_shows"`
}

// Public payloads.

type Station struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Subtitle        string         `json:"subtitle"`
	Image           []DownloadLink `json:"image"`
	URL             string         `json:"url"`
	Language        string         `json:"language"`
	SubType         string         `json:"subType"`
	ExplicitContent bool           `json:"explicitContent"`
}

type Channel struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Subtitle        string         `json:"subtitle"`
	Image           []DownloadLink `json:"image"`
	URL             string         `json:"url"`
	SubType         string         `json:"subType"`
	Type            string         `json:"type"`
	ExplicitContent bool           `json:"explicitContent"`
}

type Chart struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Subtitle        string         `json:"subtitle"`
	Image           []DownloadLink `json:"image"`
	URL             string         `json:"url"`
	Language        string         `json:"language"`
	ExplicitContent bool           `json:"explicitContent"`
}

type Shows struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Subtitle string         `json:"subtitle"`
	Image    []DownloadLink `json:"image"`
	URL      string         `json:"url"`
	Language string         `json:"language"`
}

type BrowsePlaylist struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Subtitle        string         `json:"subtitle"`
	Description     *string        `json:"description"`
	URL             string         `json:"url"`
	Image           []DownloadLink `json:"image"`
	Language        string         `json:"language"`
	Year            *int           `json:"year"`
	ExplicitContent bool           `json:"explicitContent"`
	SongCount       *int           `json:"songCount"`
	FollowerCount   *int           `json:"followerCount"`
}

type BrowseResponse struct {
	Language     string           `json:"language"`
	Stations     []Station        `json:"stations"`
	Channels     []Channel        `json:"channels"`
	Charts       []Chart          `json:"charts"`
	NewAlbums    []AlbumModel     `json:"newAlbums"`
	NewTrending  []BrowsePlaylist `json:"newTrending"`
	TopPlaylists []BrowsePlaylist `json:"topPlaylists"`
	TopShows     []Shows          `json:"topShows"`
}

// FeaturedPlaylist is one row of playlist.getFeaturedPlaylists (the
// per-language "top hits / curated" feed) — this call is language-aware
// upstream, unlike content.getCharts/getBrowseModules which some regions
// lock to their default language.
type FeaturedPlaylist struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Image           []DownloadLink `json:"image"`
	URL             string         `json:"url"`
	SongCount       *int           `json:"songCount"`
	FollowerCount   *int           `json:"followerCount"`
	ExplicitContent bool           `json:"explicitContent"`
}

type featuredPlaylistsRaw []struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Type     string `json:"type"`
	Image    string `json:"image"`
	PermaURL string `json:"perma_url"`
	Language string `json:"language"`
	Explicit string `json:"explicit_content"`
}
