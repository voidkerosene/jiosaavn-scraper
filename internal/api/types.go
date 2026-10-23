package api

import (
	"bytes"
	"encoding/json"
)

// Upstream JSON models for www.jiosaavn.com/api.php responses.
// Field types are deliberately permissive: JioSaavn mixes strings and
// numbers (and even arrays/objects in string fields) depending on the
// endpoint, so fields used as text are decoded defensively.

// StringOrNumber accepts either a JSON string or a JSON number and
// keeps the raw text, so numeric fields can be parsed afterwards.
type StringOrNumber string

func (s *StringOrNumber) String() string { return string(*s) }

func (s *StringOrNumber) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return nil
	}
	if b[0] == '"' {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*s = StringOrNumber(v)
		return nil
	}
	*s = StringOrNumber(bytes.Trim(b, " "))
	return nil
}

type ArtistMapRaw struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Image    string `json:"image"`
	Type     string `json:"type"`
	PermaURL string `json:"perma_url"`
}

// ArtistMapList tolerates the upstream quirk where artist lists are
// sometimes arrays and sometimes strings or objects.
type ArtistMapList []ArtistMapRaw

func (l *ArtistMapList) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || b[0] != '[' {
		*l = nil
		return nil
	}
	return json.Unmarshal(b, (*[]ArtistMapRaw)(l))
}

// EmptyTolerant types: several upstream fields flip between an array
// and "" / 0 depending on the endpoint, so decode defensively.

// SongList tolerates "list": []SongRaw and "list": "".
type SongList []SongRaw

func (l *SongList) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || b[0] != '[' {
		*l = nil
		return nil
	}
	return json.Unmarshal(b, (*[]SongRaw)(l))
}

// ArtistMapContainer decodes the artistMap object but ignores anything
// that isn't an object (upstream sends "" in odd responses).
type ArtistMapContainer struct {
	Primary  ArtistMapList `json:"primary_artists"`
	Featured ArtistMapList `json:"featured_artists"`
	All      ArtistMapList `json:"artists"`
}

func (c *ArtistMapContainer) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || b[0] != '{' {
		*c = ArtistMapContainer{}
		return nil
	}
	type alias ArtistMapContainer
	return json.Unmarshal(b, (*alias)(c))
}

type SongMoreInfo struct {
	Music             StringOrNumber     `json:"music"`
	AlbumID           string             `json:"album_id"`
	Album             string             `json:"album"`
	Label             string             `json:"label"`
	Has320            string             `json:"320kbps"`
	EncryptedMediaURL string             `json:"encrypted_media_url"`
	AlbumURL          string             `json:"album_url"`
	Duration          StringOrNumber     `json:"duration"`
	HasLyrics         string             `json:"has_lyrics"`
	LyricsID          string             `json:"lyrics_id"`
	Copyright         string             `json:"copyright_text"`
	ReleaseDate       string             `json:"release_date"`
	SongCount         StringOrNumber     `json:"song_count"`
	PrimaryArtists    StringOrNumber     `json:"primary_artists"`
	Singers           StringOrNumber     `json:"singers"`
	ArtistMap         ArtistMapContainer `json:"artistMap"`
}

type SongRaw struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Image           string         `json:"image"`
	Type            string         `json:"type"`
	PermaURL        string         `json:"perma_url"`
	Language        string         `json:"language"`
	Year            StringOrNumber `json:"year"`
	PlayCount       StringOrNumber `json:"play_count"`
	ExplicitContent string         `json:"explicit_content"`
	ListCount       StringOrNumber `json:"list_count"`
	List            SongList       `json:"list"`
	HeaderDesc      string         `json:"header_desc"`
	Description     string         `json:"description"`
	MoreInfo        SongMoreInfo   `json:"more_info"`
}

type SearchAllResponse struct {
	TopQuery  SearchSection `json:"topquery"`
	Songs     SearchSection `json:"songs"`
	Albums    SearchSection `json:"albums"`
	Artists   SearchSection `json:"artists"`
	Playlists SearchSection `json:"playlists"`
}

type SearchSection struct {
	Data     []SearchItemRaw `json:"data"`
	Position int             `json:"position"`
}

type SearchItemRaw struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Image           string         `json:"image"`
	Type            string         `json:"type"`
	PermaURL        string         `json:"perma_url"`
	Subtitle        string         `json:"subtitle"`
	Description     string         `json:"description"`
	Language        string         `json:"language"`
	HeaderDesc      string         `json:"header_desc"`
	ExplicitContent string         `json:"explicit_content"`
	Year            StringOrNumber `json:"year"`
	PlayCount       StringOrNumber `json:"play_count"`
	MoreInfo        SongMoreInfo   `json:"more_info"`
}

type SearchResponse struct {
	Total   int             `json:"total"`
	Start   int             `json:"start"`
	Results []SearchItemRaw `json:"results"`
}

type AlbumMoreInfo struct {
	SongCount StringOrNumber     `json:"song_count"`
	Copyright string             `json:"copyright_text"`
	ArtistMap ArtistMapContainer `json:"artistMap"`
}

type AlbumRaw struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Image           string         `json:"image"`
	Type            string         `json:"type"`
	PermaURL        string         `json:"perma_url"`
	Language        string         `json:"language"`
	Year            StringOrNumber `json:"year"`
	PlayCount       StringOrNumber `json:"play_count"`
	ExplicitContent string         `json:"explicit_content"`
	HeaderDesc      string         `json:"header_desc"`
	List            SongList       `json:"list"`
	MoreInfo        AlbumMoreInfo  `json:"more_info"`
}

type PlaylistRaw struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Image           string         `json:"image"`
	Type            string         `json:"type"`
	PermaURL        string         `json:"perma_url"`
	Language        string         `json:"language"`
	Year            StringOrNumber `json:"year"`
	PlayCount       StringOrNumber `json:"play_count"`
	ExplicitContent string         `json:"explicit_content"`
	HeaderDesc      string         `json:"header_desc"`
	List            SongList       `json:"list"`
	ListCount       StringOrNumber `json:"list_count"`
	MoreInfo        struct {
		Artists []ArtistMapRaw `json:"artists"`
	} `json:"more_info"`
}

type SongByLinkResponse struct {
	Songs []SongRaw `json:"songs"`
}

// song.getDetails (by pids) wraps the list in a "songs" object;
// song-by-link (webapi.get + type=song) returns the same shape.
type SongDetailsResponse struct {
	Songs []SongRaw `json:"songs"`
}

type ArtistDetailsResponse struct {
	ArtistID           string          `json:"artistId"`
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Image              string          `json:"image"`
	Type               string          `json:"type"`
	PermaURL           string          `json:"perma_url"`
	FollowerCount      StringOrNumber  `json:"follower_count"`
	FanCount           string          `json:"fan_count"`
	IsVerified         bool            `json:"isVerified"`
	DominantLanguage   string          `json:"dominantLanguage"`
	DominantType       string          `json:"dominantType"`
	Dob                string          `json:"dob"`
	Fb                 string          `json:"fb"`
	Twitter            string          `json:"twitter"`
	Wiki               string          `json:"wiki"`
	Bio                json.RawMessage `json:"bio"`
	AvailableLanguages []string        `json:"availableLanguages"`
	IsRadioPresent     bool            `json:"isRadioPresent"`
	URLs               struct {
		Overview string `json:"overview"`
	} `json:"urls"`
	TopSongs  []SongRaw  `json:"topSongs"`
	TopAlbums []AlbumRaw `json:"topAlbums"`
	Singles   []SongRaw  `json:"singles"`
	Similar   []struct {
		ID             string          `json:"id"`
		Name           string          `json:"name"`
		ImageURL       string          `json:"image_url"`
		PermaURL       string          `json:"perma_url"`
		Type           string          `json:"type"`
		DominantType   string          `json:"dominantType"`
		Aka            string          `json:"aka"`
		Bio            json.RawMessage `json:"bio"`
		Languages      json.RawMessage `json:"languages"`
		Similar        json.RawMessage `json:"similar"`
		Wiki           string          `json:"wiki"`
		Dob            string          `json:"dob"`
		Fb             string          `json:"fb"`
		Twitter        string          `json:"twitter"`
		IsRadioPresent bool            `json:"isRadioPresent"`
	} `json:"similarArtists"`
}

type SongStationResponse struct {
	StationID string `json:"stationid"`
}

// TrendingListResponse is a bare JSON array of trending entities.
type TrendingListResponse []TrendingItemRaw

type ArtistSongsResponse struct {
	TopSongs struct {
		Total int       `json:"total"`
		Songs []SongRaw `json:"songs"`
	} `json:"topSongs"`
}

type ArtistAlbumsResponse struct {
	TopAlbums struct {
		Total  int        `json:"total"`
		Albums []AlbumRaw `json:"albums"`
	} `json:"topAlbums"`
}

type TrendingItemRaw struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Image           string         `json:"image"`
	Type            string         `json:"type"`
	PermaURL        string         `json:"perma_url"`
	Language        string         `json:"language"`
	Year            StringOrNumber `json:"year"`
	ExplicitContent string         `json:"explicit_content"`
	ListCount       StringOrNumber `json:"list_count"`
	PlayCount       StringOrNumber `json:"play_count"`
	MoreInfo        SongMoreInfo   `json:"more_info"`
}
