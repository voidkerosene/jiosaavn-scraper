package api

import (
	"context"
	"strconv"
	"strings"
)

// Browse mirrors the JioSaavn home page (jiosaavn.com equivalent of
// "top hits trending / charts / new releases / stations / mood-genre
// channels") for a content language. One upstream call per language,
// cached for 5 minutes by the server layer.
func Browse(ctx context.Context, language, section string) (BrowseResponse, error) {
	language = normalizeLanguage(language)
	if language == "" {
		language = "hindi"
	}

	sec := BrowseSection(section)
	if section == "" {
		sec = SectionAll
	}
	if !sectionAllowed(sec) {
		return BrowseResponse{}, errBadInput("invalid section: use all, stations, channels, charts, newAlbums, newTrending, topPlaylists or topShows")
	}

	var raw browseModulesRaw
	if err := fetch(ctx, callBrowseModules, map[string]string{
		"entity_language": language,
	}, &raw); err != nil {
		return BrowseResponse{}, errNotFound("browse modules not found for this language")
	}

	resp := BrowseResponse{}
	want := func(s BrowseSection) bool {
		return sec == SectionAll || sec == s
	}

	if want(SectionStations) {
		resp.Stations = make([]Station, 0, len(raw.Radio.FeaturedStations))
		for _, st := range raw.Radio.FeaturedStations {
			resp.Stations = append(resp.Stations, Station{
				ID:              st.ID,
				Name:            sanitize(st.Title),
				Subtitle:        sanitize(st.Subtitle),
				Image:           CreateImageLinks(st.Image),
				URL:             st.PermaURL,
				Language:        st.Language,
				SubType:         st.MoreInfo.SubType,
				ExplicitContent: boolVal(st.ExplicitContent),
			})
		}
	}
	if want(SectionChannels) {
		resp.Channels = make([]Channel, 0, len(raw.BrowseDiscover))
		for _, ch := range raw.BrowseDiscover {
			resp.Channels = append(resp.Channels, Channel{
				ID:              ch.ID,
				Name:            sanitize(ch.Title),
				Subtitle:        sanitize(ch.Subtitle),
				Image:           CreateImageLinks(ch.Image),
				URL:             ch.PermaURL,
				SubType:         ch.MoreInfo.SubType,
				Type:            ch.Type,
				ExplicitContent: boolVal(ch.ExplicitContent),
			})
		}
	}
	if want(SectionCharts) {
		resp.Charts = make([]Chart, 0, len(raw.Charts))
		for _, c := range raw.Charts {
			resp.Charts = append(resp.Charts, Chart{
				ID:              c.ID,
				Name:            sanitize(c.Title),
				Subtitle:        sanitize(c.Subtitle),
				Image:           CreateImageLinks(c.Image),
				URL:             c.PermaURL,
				Language:        c.Language,
				ExplicitContent: boolVal(c.ExplicitContent),
			})
		}
	}
	if want(SectionNewAlbums) {
		resp.NewAlbums = make([]AlbumModel, 0, len(raw.NewAlbums))
		for _, a := range raw.NewAlbums {
			if a.Type != "album" || a.ID == "" {
				continue
			}
			resp.NewAlbums = append(resp.NewAlbums, albumModel(a))
		}
	}
	if want(SectionNewTrending) {
		resp.NewTrending = playlistish(raw.NewTrending)
	}
	if want(SectionTopPlaylists) {
		resp.TopPlaylists = playlistish(raw.TopPlaylists)
	}
	if want(SectionTopShows) {
		resp.TopShows = make([]Shows, 0, len(raw.TopShows.Shows))
		for _, sh := range raw.TopShows.Shows {
			resp.TopShows = append(resp.TopShows, Shows{
				ID:       sh.ID,
				Name:     sanitize(sh.Title),
				Subtitle: sanitize(sh.Subtitle),
				Image:    CreateImageLinks(sh.Image),
				URL:      sh.PermaURL,
				Language: sh.Language,
			})
		}
	}
	resp.Language = language
	return resp, nil
}

// FeaturedPlaylists is the verified per-language curated/top-hits
// playlist feed (playlist.getFeaturedPlaylists). Upstream supports
// offsets and huge sizes; we clamp to sane limits.
func FeaturedPlaylists(ctx context.Context, language string, page, limit int) (SearchPlaylistsResponse, error) {
	language = normalizeLanguage(language)
	if language == "" {
		language = "hindi"
	}
	limit = clampLimit(limit, 10, 250)

	var raw featuredPlaylistsRaw
	if err := fetch(ctx, callFeaturedPlaylists, map[string]string{
		"language": language,
		"offset":   strconv.Itoa(nonNegative(page)),
		"size":     strconv.Itoa(limit),
	}, &raw); err != nil {
		return SearchPlaylistsResponse{}, errNotFound("featured playlists not found for this language")
	}

	results := make([]SearchPlaylistItem, 0, len(raw))
	for _, p := range raw {
		name := sanitize(p.Title)
		if p.ID == "" || name == "" {
			continue
		}
		results = append(results, SearchPlaylistItem{
			ID:              p.ID,
			Name:            name,
			Type:            p.Type,
			Image:           CreateImageLinks(p.Image),
			URL:             p.PermaURL,
			SongCount:       songCountFromSubtitle(p.Subtitle),
			Language:        p.Language,
			ExplicitContent: boolVal(p.Explicit),
		})
	}
	return SearchPlaylistsResponse{Total: len(results), Start: page, Results: results}, nil
}

// songCountFromSubtitle extracts the count from subtitles like
// "23 Songs" or "1.2K Songs" (upstream abbreviates large counts).
func songCountFromSubtitle(sub string) *int {
	sub = strings.ToLower(strings.TrimSpace(sub))
	i := strings.IndexAny(sub, "0123456789")
	if i < 0 {
		return nil
	}
	end := i
	for end < len(sub) && ((sub[end] >= '0' && sub[end] <= '9') || sub[end] == '.') {
		end++
	}
	f, err := strconv.ParseFloat(sub[i:end], 64)
	if err != nil {
		return nil
	}
	if end < len(sub) {
		switch sub[end] {
		case 'k':
			f *= 1_000
		case 'm':
			f *= 1_000_000
		case 'b':
			f *= 1_000_000_000
		}
	}
	v := int(f)
	return &v
}

func playlistish(items []BrowsePlaylistRaw) []BrowsePlaylist {
	out := make([]BrowsePlaylist, 0, len(items))
	for _, p := range items {
		if p.ID == "" {
			continue
		}
		out = append(out, BrowsePlaylist{
			ID:              p.ID,
			Name:            sanitize(p.Title),
			Subtitle:        sanitize(p.Subtitle),
			Description:     sPtr(p.HeaderDesc),
			URL:             p.PermaURL,
			Image:           CreateImageLinks(p.Image),
			Language:        p.Language,
			Year:            intPtr(p.Year),
			ExplicitContent: boolVal(p.ExplicitContent),
			SongCount:       intPtr(p.MoreInfo.SongCount),
			FollowerCount:   intPtr(p.MoreInfo.FollowerCount),
		})
	}
	return out
}

// normalizeLanguage maps common casing/aliases to the lowercase codes
// JioSaavn expects (entity_language).
func normalizeLanguage(l string) string {
	return strings.ToLower(strings.TrimSpace(l))
}
