package api

import "encoding/json"

// Public response payloads returned by this API.

type DownloadLink struct {
	Quality string `json:"quality"`
	URL     string `json:"url"`
}

type SongModel struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	Year            *string          `json:"year"`
	ReleaseDate     *string          `json:"releaseDate"`
	Duration        *int             `json:"duration"`
	Label           *string          `json:"label"`
	ExplicitContent bool             `json:"explicitContent"`
	PlayCount       *int             `json:"playCount"`
	Language        string           `json:"language"`
	HasLyrics       bool             `json:"hasLyrics"`
	LyricsID        *string          `json:"lyricsId"`
	URL             string           `json:"url"`
	Copyright       *string          `json:"copyright"`
	Album           SongAlbumModel   `json:"album"`
	Artists         SongArtistsModel `json:"artists"`
	Image           []DownloadLink   `json:"image"`
	DownloadURL     []DownloadLink   `json:"downloadUrl"`
}

type SongAlbumModel struct {
	ID   *string `json:"id"`
	Name *string `json:"name"`
	URL  *string `json:"url"`
}

type SongArtistsModel struct {
	Primary  []ArtistMapModel `json:"primary"`
	Featured []ArtistMapModel `json:"featured"`
	All      []ArtistMapModel `json:"all"`
}

type ArtistMapModel struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Role  string         `json:"role"`
	Image []DownloadLink `json:"image"`
	Type  string         `json:"type"`
	URL   string         `json:"url"`
}

type AlbumModel struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Description     *string          `json:"description"`
	Year            *int             `json:"year"`
	Type            string           `json:"type"`
	PlayCount       *int             `json:"playCount"`
	Language        string           `json:"language"`
	ExplicitContent bool             `json:"explicitContent"`
	URL             string           `json:"url"`
	SongCount       *int             `json:"songCount"`
	Artists         SongArtistsModel `json:"artists"`
	Image           []DownloadLink   `json:"image"`
	Songs           []SongModel      `json:"songs"`
}

type PlaylistModel struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Description     *string          `json:"description"`
	Year            *int             `json:"year"`
	Type            string           `json:"type"`
	PlayCount       *int             `json:"playCount"`
	Language        string           `json:"language"`
	ExplicitContent bool             `json:"explicitContent"`
	URL             string           `json:"url"`
	SongCount       *int             `json:"songCount"`
	Artists         []ArtistMapModel `json:"artists"`
	Image           []DownloadLink   `json:"image"`
	Songs           []SongModel      `json:"songs"`
}

type SearchSectionModel[T any] struct {
	Results  []T `json:"results"`
	Position int `json:"position"`
}

// GlobalSearchResponse is returned by /api/search (autocomplete).
type GlobalSearchResponse struct {
	TopQuery  SearchSectionModel[GlobalSearchItem] `json:"topQuery"`
	Songs     SearchSectionModel[GlobalSearchItem] `json:"songs"`
	Albums    SearchSectionModel[GlobalSearchItem] `json:"albums"`
	Artists   SearchSectionModel[GlobalSearchItem] `json:"artists"`
	Playlists SearchSectionModel[GlobalSearchItem] `json:"playlists"`
}

type GlobalSearchItem struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Image          []DownloadLink `json:"image"`
	Album          *string        `json:"album"`
	URL            *string        `json:"url"`
	Type           string         `json:"type"`
	Description    *string        `json:"description"`
	PrimaryArtists *string        `json:"primaryArtists"`
	Singers        *string        `json:"singers"`
	Language       *string        `json:"language"`
	Artist         *string        `json:"artist"`
	Year           *string        `json:"year"`
}

type SearchSongsResponse struct {
	Total   int         `json:"total"`
	Start   int         `json:"start"`
	Results []SongModel `json:"results"`
}

type SearchAlbumsResponse struct {
	Total   int          `json:"total"`
	Start   int          `json:"start"`
	Results []AlbumModel `json:"results"`
}

type SearchArtistsResponse struct {
	Total   int              `json:"total"`
	Start   int              `json:"start"`
	Results []ArtistMapModel `json:"results"`
}

type SearchPlaylistsResponse struct {
	Total   int                  `json:"total"`
	Start   int                  `json:"start"`
	Results []SearchPlaylistItem `json:"results"`
}

type SearchPlaylistItem struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Image           []DownloadLink `json:"image"`
	URL             string         `json:"url"`
	SongCount       *int           `json:"songCount"`
	Language        string         `json:"language"`
	ExplicitContent bool           `json:"explicitContent"`
}

type ArtistModel struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	URL                string          `json:"url"`
	Type               string          `json:"type"`
	Image              []DownloadLink  `json:"image"`
	FollowerCount      *int            `json:"followerCount"`
	FanCount           *string         `json:"fanCount"`
	IsVerified         *bool           `json:"isVerified"`
	DominantLanguage   *string         `json:"dominantLanguage"`
	DominantType       *string         `json:"dominantType"`
	Bio                json.RawMessage `json:"bio"`
	Dob                *string         `json:"dob"`
	Fb                 *string         `json:"fb"`
	Twitter            *string         `json:"twitter"`
	Wiki               *string         `json:"wiki"`
	AvailableLanguages []string        `json:"availableLanguages"`
	IsRadioPresent     bool            `json:"isRadioPresent"`
	TopSongs           []SongModel     `json:"topSongs"`
	TopAlbums          []AlbumModel    `json:"topAlbums"`
	Singles            []SongModel     `json:"singles"`
	SimilarArtists     []SimilarArtist `json:"similarArtists"`
}

type SimilarArtist struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	URL            *string         `json:"url"`
	Image          []DownloadLink  `json:"image"`
	Languages      json.RawMessage `json:"languages"`
	Wiki           *string         `json:"wiki"`
	Dob            *string         `json:"dob"`
	Fb             *string         `json:"fb"`
	Twitter        *string         `json:"twitter"`
	IsRadioPresent bool            `json:"isRadioPresent"`
	Type           string          `json:"type"`
	DominantType   string          `json:"dominantType"`
	Aka            string          `json:"aka"`
	Bio            json.RawMessage `json:"bio"`
	Similar        json.RawMessage `json:"similarArtists"`
}

type TrendingItem struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Image           []DownloadLink   `json:"image"`
	Type            string           `json:"type"`
	URL             string           `json:"url"`
	Language        string           `json:"language"`
	Year            *string          `json:"year"`
	ExplicitContent bool             `json:"explicitContent"`
	SongCount       *int             `json:"songCount"`
	PlayCount       *int             `json:"playCount"`
	Album           *string          `json:"album"`
	Singers         *string          `json:"singers"`
	PrimaryArtists  *string          `json:"primaryArtists"`
	Artists         []ArtistMapModel `json:"artists"`
}

type TrendingResponse struct {
	Results []TrendingItem `json:"results"`
}
