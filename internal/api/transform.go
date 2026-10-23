package api

import (
	"encoding/json"
	"html"
	"strconv"
	"strings"
)

// sanitize cleans strings coming from upstream: unescapes HTML entities
// (&quot; etc.) and drops JioSaavn's occasional server-template junk
// like "<%!1=2%>".
func sanitize(s string) string {
	if strings.Contains(s, "<%") {
		return ""
	}
	out := html.UnescapeString(s)
	return strings.TrimSpace(out)
}

// strPtr mirrors the original API: empty upstream values become null.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	v := sanitize(s)
	if v == "" {
		return nil
	}
	return &v
}

// flexStrPtr parses a StringOrNumber into a nullable string.
func flexStrPtr(s StringOrNumber) *string {
	return strPtr(s.String())
}

// intPtr parses a numeric string into a nullable int ("0" stays 0,
// "" becomes null, garbage like "1.0e7" is parsed too).
func intPtr(s StringOrNumber) *int {
	t := strings.TrimSpace(s.String())
	if t == "" {
		return nil
	}
	if n, err := strconv.ParseInt(t, 10, 32); err == nil {
		v := int(n)
		return &v
	}
	if f, err := strconv.ParseFloat(t, 32); err == nil {
		v := int(f)
		return &v
	}
	return nil
}

func boolPtr(b bool) *bool { return &b }

func boolVal(s string) bool { return s == "1" }

func sPtr(s string) *string {
	if s == "" {
		return nil
	}
	v := sanitize(s)
	return &v
}

// jsonEmbed embeds an upstream JSON field (bio, languages, ...) into
// the output. When upstream sends it twice-encoded ("[{...}]"), it
// unwraps one level so the result is real JSON, not a quoted string.
func jsonEmbed(raw json.RawMessage) json.RawMessage {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return nil
	}
	if strings.HasPrefix(s, `"`) {
		var inner string
		if err := json.Unmarshal([]byte(s), &inner); err != nil {
			return nil
		}
		inner = strings.TrimSpace(inner)
		if inner == "" {
			return nil
		}
		return json.RawMessage(inner)
	}
	return json.RawMessage(s)
}

func artistMapModel(m ArtistMapRaw) ArtistMapModel {
	return ArtistMapModel{
		ID:    m.ID,
		Name:  sanitize(m.Name),
		Role:  m.Role,
		Image: CreateImageLinks(m.Image),
		Type:  m.Type,
		URL:   m.PermaURL,
	}
}

func artistMapModels(list ArtistMapList) []ArtistMapModel {
	if len(list) == 0 {
		return []ArtistMapModel{}
	}
	out := make([]ArtistMapModel, 0, len(list))
	for _, m := range list {
		out = append(out, artistMapModel(m))
	}
	return out
}

func songModel(s SongRaw) SongModel {
	more := s.MoreInfo
	links := CreateDownloadLinks(more.EncryptedMediaURL)
	// Only advertise 320kbps when the master truly has it; the
	// replacement link would 404 otherwise.
	if more.Has320 != "true" {
		for i, l := range links {
			if strings.HasSuffix(l.URL, "_320.mp4") {
				links = append(links[:i], links[i+1:]...)
				break
			}
		}
	}
	return SongModel{
		ID:              s.ID,
		Name:            sanitize(s.Title),
		Type:            s.Type,
		Year:            flexStrPtr(s.Year),
		ReleaseDate:     sPtr(more.ReleaseDate),
		Duration:        intPtr(more.Duration),
		Label:           sPtr(more.Label),
		ExplicitContent: boolVal(s.ExplicitContent),
		PlayCount:       intPtr(s.PlayCount),
		Language:        s.Language,
		HasLyrics:       more.HasLyrics == "true",
		LyricsID:        sPtr(more.LyricsID),
		URL:             s.PermaURL,
		Copyright:       sPtr(more.Copyright),
		Album: SongAlbumModel{
			ID:   sPtr(more.AlbumID),
			Name: sPtr(more.Album),
			URL:  sPtr(more.AlbumURL),
		},
		Artists: SongArtistsModel{
			Primary:  artistMapModels(more.ArtistMap.Primary),
			Featured: artistMapModels(more.ArtistMap.Featured),
			All:      artistMapModels(more.ArtistMap.All),
		},
		Image:       CreateImageLinks(s.Image),
		DownloadURL: links,
	}
}

func songModels(list []SongRaw) []SongModel {
	out := make([]SongModel, 0, len(list))
	for _, s := range list {
		out = append(out, songModel(s))
	}
	return out
}

func albumModel(a AlbumRaw) AlbumModel {
	songs := songModels(a.List)
	if songs == nil {
		songs = []SongModel{}
	}
	return AlbumModel{
		ID:              a.ID,
		Name:            sanitize(a.Title),
		Description:     sPtr(a.HeaderDesc),
		Year:            intPtr(a.Year),
		Type:            a.Type,
		PlayCount:       intPtr(a.PlayCount),
		Language:        a.Language,
		ExplicitContent: boolVal(a.ExplicitContent),
		URL:             a.PermaURL,
		SongCount:       intPtr(a.MoreInfo.SongCount),
		Artists: SongArtistsModel{
			Primary:  artistMapModels(a.MoreInfo.ArtistMap.Primary),
			Featured: artistMapModels(a.MoreInfo.ArtistMap.Featured),
			All:      artistMapModels(a.MoreInfo.ArtistMap.All),
		},
		Image: CreateImageLinks(a.Image),
		Songs: songs,
	}
}

func albumModels(list []AlbumRaw) []AlbumModel {
	out := make([]AlbumModel, 0, len(list))
	for _, a := range list {
		out = append(out, albumModel(a))
	}
	return out
}

func playlistModel(p PlaylistRaw) PlaylistModel {
	songs := []SongModel{}
	for _, s := range p.List {
		songs = append(songs, songModel(s))
	}
	artists := []ArtistMapModel{}
	for _, a := range p.MoreInfo.Artists {
		artists = append(artists, artistMapModel(a))
	}
	return PlaylistModel{
		ID:              p.ID,
		Name:            sanitize(p.Title),
		Description:     sPtr(p.HeaderDesc),
		Year:            intPtr(p.Year),
		Type:            p.Type,
		PlayCount:       intPtr(p.PlayCount),
		Language:        p.Language,
		ExplicitContent: boolVal(p.ExplicitContent),
		URL:             p.PermaURL,
		SongCount:       intPtr(p.ListCount),
		Artists:         artists,
		Image:           CreateImageLinks(p.Image),
		Songs:           songs,
	}
}

func artistModel(a ArtistDetailsResponse) ArtistModel {
	id := a.ArtistID
	if id == "" {
		id = a.ID
	}
	userURL := a.URLs.Overview
	if userURL == "" {
		userURL = a.PermaURL
	}
	topSongs := []SongModel{}
	for _, s := range a.TopSongs {
		topSongs = append(topSongs, songModel(s))
	}
	topAlbums := []AlbumModel{}
	for _, al := range a.TopAlbums {
		topAlbums = append(topAlbums, albumModel(al))
	}
	singles := []SongModel{}
	for _, s := range a.Singles {
		singles = append(singles, songModel(s))
	}
	similar := []SimilarArtist{}
	for _, s := range a.Similar {
		similar = append(similar, SimilarArtist{
			ID:             s.ID,
			Name:           sanitize(s.Name),
			URL:            sPtr(s.PermaURL),
			Image:          CreateImageLinks(s.ImageURL),
			Languages:      jsonEmbed(s.Languages),
			Wiki:           sPtr(s.Wiki),
			Dob:            sPtr(s.Dob),
			Fb:             sPtr(s.Fb),
			Twitter:        sPtr(s.Twitter),
			IsRadioPresent: s.IsRadioPresent,
			Type:           s.Type,
			DominantType:   s.DominantType,
			Aka:            sanitize(s.Aka),
			Bio:            jsonEmbed(s.Bio),
			Similar:        jsonEmbed(s.Similar),
		})
	}
	return ArtistModel{
		ID:                 id,
		Name:               sanitize(a.Name),
		URL:                userURL,
		Type:               a.Type,
		Image:              CreateImageLinks(a.Image),
		FollowerCount:      intPtr(a.FollowerCount),
		FanCount:           sPtr(a.FanCount),
		IsVerified:         boolPtr(a.IsVerified),
		DominantLanguage:   sPtr(a.DominantLanguage),
		DominantType:       sPtr(a.DominantType),
		Bio:                jsonEmbed(a.Bio),
		Dob:                sPtr(a.Dob),
		Fb:                 sPtr(a.Fb),
		Twitter:            sPtr(a.Twitter),
		Wiki:               sPtr(a.Wiki),
		AvailableLanguages: a.AvailableLanguages,
		IsRadioPresent:     a.IsRadioPresent,
		TopSongs:           topSongs,
		TopAlbums:          topAlbums,
		Singles:            singles,
		SimilarArtists:     similar,
	}
}

func trendingItemModel(t TrendingItemRaw) TrendingItem {
	more := t.MoreInfo
	item := TrendingItem{
		ID:              t.ID,
		Title:           sanitize(t.Title),
		Image:           CreateImageLinks(t.Image),
		Type:            t.Type,
		URL:             t.PermaURL,
		Language:        t.Language,
		Year:            flexStrPtr(t.Year),
		ExplicitContent: boolVal(t.ExplicitContent),
		SongCount:       intPtr(more.SongCount),
		PlayCount:       intPtr(t.PlayCount),
		PrimaryArtists:  flexStrPtr(more.PrimaryArtists),
		Singers:         flexStrPtr(more.Singers),
	}
	if more.Album != "" {
		item.Album = sPtr(more.Album)
	}
	if len(more.ArtistMap.All) > 0 {
		item.Artists = artistMapModels(more.ArtistMap.All)
	} else if len(more.ArtistMap.Primary) > 0 {
		item.Artists = artistMapModels(more.ArtistMap.Primary)
	} else {
		item.Artists = []ArtistMapModel{}
	}
	return item
}
