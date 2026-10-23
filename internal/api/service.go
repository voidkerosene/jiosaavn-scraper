package api

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// HTTPError maps failures to a status code + message.
type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

// ErrNotFound is the sentinel the client returns on upstream 404s.
var ErrNotFound = &HTTPError{Code: 404, Message: "not found"}

func errBadInput(msg string) error { return &HTTPError{Code: 400, Message: msg} }
func errNotFound(msg string) error { return &HTTPError{Code: 404, Message: msg} }
func errInternal(msg string) error { return &HTTPError{Code: 500, Message: msg} }

// clampLimit keeps pagination inside sane bounds; upstream fails hard
// on huge page numbers or zero limits.
func clampLimit(n, def, maxCount int) int {
	if n <= 0 {
		return def
	}
	if n > maxCount {
		return maxCount
	}
	return n
}

func nonNegative(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

func sortOrDefault(v, def string) string {
	switch v {
	case "popularity", "latest", "alphabetical":
		return v
	case "asc", "desc":
		return v
	}
	return def
}

// lastSegment extracts the trailing token used in JioSaavn perma links
// (e.g. https://www.jiosaavn.com/album/x/ITIyo-GDr7A_ -> "ITIyo-GDr7A_").
func lastSegment(rawLink string) (string, error) {
	link := strings.TrimSpace(rawLink)
	if link == "" || !strings.Contains(link, "jiosaavn.com/") {
		return "", errBadInput("invalid JioSaavn link")
	}
	token := link[strings.LastIndex(link, "jiosaavn.com/")+len("jiosaavn.com/"):]
	if idx := strings.LastIndex(token, "/"); idx != -1 && idx < len(token)-1 {
		token = token[idx+1:]
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errBadInput("invalid JioSaavn link")
	}
	return token, nil
}

const searchDefaultLimit = 10
const searchMaxLimit = 100

// SearchSongs searches songs with pagination. Results are filtered to
// actual song entities and de-duplicated by ID (upstream sometimes
// injects a "top result" item of a different type).
func SearchSongs(ctx context.Context, query string, page, limit int) (SearchSongsResponse, error) {
	data, err := searchEntity(ctx, callSearchSongs, "song", query, page, limit)
	if err != nil {
		return SearchSongsResponse{}, err
	}
	seen := map[string]bool{}
	results := make([]SongModel, 0, len(data.Results))
	for _, raw := range data.Results {
		if seen[raw.ID] {
			continue
		}
		seen[raw.ID] = true
		results = append(results, songModel(SongRaw{
			ID: raw.ID, Title: raw.Title, Image: raw.Image, Type: raw.Type,
			PermaURL: raw.PermaURL, Language: raw.Language, Year: raw.Year,
			PlayCount: raw.PlayCount, ExplicitContent: raw.ExplicitContent,
			MoreInfo: raw.MoreInfo,
		}))
	}
	return SearchSongsResponse{Total: data.Total, Start: data.Start, Results: results}, nil
}

func SearchAlbums(ctx context.Context, query string, page, limit int) (SearchAlbumsResponse, error) {
	data, err := searchEntity(ctx, callSearchAlbums, "album", query, page, limit)
	if err != nil {
		return SearchAlbumsResponse{}, err
	}
	seen := map[string]bool{}
	results := make([]AlbumModel, 0, len(data.Results))
	for _, raw := range data.Results {
		if seen[raw.ID] {
			continue
		}
		seen[raw.ID] = true
		results = append(results, albumModel(AlbumRaw{
			ID: raw.ID, Title: raw.Title, Image: raw.Image, Type: raw.Type,
			PermaURL: raw.PermaURL, Language: raw.Language, Year: raw.Year,
			PlayCount: raw.PlayCount, ExplicitContent: raw.ExplicitContent,
			HeaderDesc: raw.HeaderDesc,
			MoreInfo: AlbumMoreInfo{
				SongCount: raw.MoreInfo.SongCount,
				ArtistMap: raw.MoreInfo.ArtistMap,
			},
		}))
	}
	return SearchAlbumsResponse{Total: data.Total, Start: data.Start, Results: results}, nil
}

func SearchArtists(ctx context.Context, query string, page, limit int) (SearchArtistsResponse, error) {
	data, err := searchEntity(ctx, callSearchArtists, "artist", query, page, limit)
	if err != nil {
		return SearchArtistsResponse{}, err
	}
	seen := map[string]bool{}
	results := make([]ArtistMapModel, 0, len(data.Results))
	for _, raw := range data.Results {
		if seen[raw.ID] {
			continue
		}
		seen[raw.ID] = true
		results = append(results, artistMapModel(ArtistMapRaw{
			ID: raw.ID, Name: raw.Title, Image: raw.Image, Type: raw.Type,
			PermaURL: raw.PermaURL,
		}))
	}
	return SearchArtistsResponse{Total: data.Total, Start: data.Start, Results: results}, nil
}

func SearchPlaylists(ctx context.Context, query string, page, limit int) (SearchPlaylistsResponse, error) {
	data, err := searchEntity(ctx, callSearchPlaylists, "playlist", query, page, limit)
	if err != nil {
		return SearchPlaylistsResponse{}, err
	}
	seen := map[string]bool{}
	results := make([]SearchPlaylistItem, 0, len(data.Results))
	for _, raw := range data.Results {
		if seen[raw.ID] {
			continue
		}
		seen[raw.ID] = true
		results = append(results, SearchPlaylistItem{
			ID:              raw.ID,
			Name:            sanitize(raw.Title),
			Type:            raw.Type,
			Image:           CreateImageLinks(raw.Image),
			URL:             raw.PermaURL,
			SongCount:       intPtr(raw.MoreInfo.SongCount),
			Language:        raw.Language,
			ExplicitContent: boolVal(raw.ExplicitContent),
		})
	}
	return SearchPlaylistsResponse{Total: data.Total, Start: data.Start, Results: results}, nil
}

// searchEntity performs a paginated upstream search and filters out the
// junk the original API returned in its results: entries with a
// different entity type than requested, duplicates, and empty IDs.
func searchEntity(ctx context.Context, call, kind, query string, page, limit int) (SearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return SearchResponse{}, errBadInput("search query is required")
	}

	var data SearchResponse
	if err := fetch(ctx, call, map[string]string{
		"q": query,
		"p": strconv.Itoa(nonNegative(page)),
		"n": strconv.Itoa(clampLimit(limit, searchDefaultLimit, searchMaxLimit)),
	}, &data); err != nil {
		return SearchResponse{}, err
	}

	seen := map[string]bool{}
	out := make([]SearchItemRaw, 0, len(data.Results))
	for _, it := range data.Results {
		if it.Type != kind || it.ID == "" || seen[it.ID] {
			continue
		}
		seen[it.ID] = true
		out = append(out, it)
	}
	data.Results = out
	return data, nil
}

// SearchAll performs a global search across songs, albums, artists,
// playlists, and the top query autocomplete section. Upstream returns a
// bare "[]" (instead of the normal object) in some regions, so the
// response is decoded shape-agnostically and an empty result is
// reported cleanly instead of a crash.
func SearchAll(ctx context.Context, query string) (GlobalSearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return GlobalSearchResponse{}, errBadInput("search query is required")
	}

	var raw json.RawMessage
	if err := fetch(ctx, callSearchAll, map[string]string{"q": query}, &raw); err != nil {
		return GlobalSearchResponse{}, err
	}

	section := func(s SearchSection) SearchSectionModel[GlobalSearchItem] {
		items := make([]GlobalSearchItem, 0, len(s.Data))
		seen := map[string]bool{}
		for _, d := range s.Data {
			if d.ID == "" || seen[d.ID+"/"+d.Type] {
				continue
			}
			seen[d.ID+"/"+d.Type] = true
			items = append(items, GlobalSearchItem{
				ID:             d.ID,
				Title:          sanitize(d.Title),
				Image:          CreateImageLinks(d.Image),
				Album:          sPtr(d.MoreInfo.Album),
				URL:            sPtr(d.PermaURL),
				Type:           d.Type,
				Description:    sPtr(d.Description),
				PrimaryArtists: flexStrPtr(d.MoreInfo.PrimaryArtists),
				Singers:        flexStrPtr(d.MoreInfo.Singers),
				Language:       sPtr(d.Language),
				Artist:         flexStrPtr(d.MoreInfo.Music),
				Year:           flexStrPtr(d.Year),
			})
		}
		return SearchSectionModel[GlobalSearchItem]{Results: items, Position: s.Position}
	}

	empty := GlobalSearchResponse{
		TopQuery:  section(SearchSection{}),
		Songs:     section(SearchSection{}),
		Albums:    section(SearchSection{}),
		Artists:   section(SearchSection{}),
		Playlists: section(SearchSection{}),
	}
	if len(raw) == 0 || raw[0] == '[' {
		// Upstream returned a bare array (observed in some regions):
		// report a clean empty result instead of a decode crash.
		return empty, nil
	}

	var data SearchAllResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return empty, nil
	}

	return GlobalSearchResponse{
		TopQuery:  section(data.TopQuery),
		Songs:     section(data.Songs),
		Albums:    section(data.Albums),
		Artists:   section(data.Artists),
		Playlists: section(data.Playlists),
	}, nil
}

// GetSong returns one or more songs by comma-separated IDs, or by a
// JioSaavn song link.
func GetSong(ctx context.Context, ids, link string) ([]SongModel, error) {
	if strings.TrimSpace(link) != "" {
		token, err := lastSegment(link)
		if err != nil {
			return nil, err
		}
		var data SongByLinkResponse
		if err := fetch(ctx, callWebAPI, map[string]string{
			"token": token, "type": "song",
		}, &data); err != nil {
			return nil, errNotFound("song not found")
		}
		if len(data.Songs) == 0 {
			return nil, errNotFound("song not found")
		}
		return songModels(data.Songs), nil
	}

	ids = strings.Trim(ids, " ,")
	if ids == "" {
		return nil, errBadInput("either song ids or link is required")
	}

	var data SongDetailsResponse
	if err := fetch(ctx, callSongDetails, map[string]string{"pids": ids}, &data); err != nil {
		return nil, errNotFound("song not found")
	}
	if len(data.Songs) == 0 {
		return nil, errNotFound("song not found")
	}
	return songModels(data.Songs), nil
}

// GetSongSuggestions creates a radio station for the song and returns
// similar songs for infinite playback. Upstream's radio backend is
// flaky (it often answers "No new song found for current radio.");
// we retry once and then return a clean 404 instead of crashing the
// response like the original does on this exact payload.
func GetSongSuggestions(ctx context.Context, songID string, limit int) ([]SongModel, error) {
	if strings.TrimSpace(songID) == "" {
		return nil, errBadInput("song id is required")
	}
	limit = clampLimit(limit, searchDefaultLimit, searchMaxLimit)

	var station SongStationResponse
	if err := fetch(ctx, callCreateStation, map[string]string{
		"entity_id": urlencodedList(songID), "entity_type": "queue",
	}, &station); err != nil || station.StationID == "" {
		// Upstream radio creation is geo-flaky; report as empty
		// suggestions rather than a server error.
		return nil, errNotFound("no suggestions found for the given song")
	}

	for attempt := 0; attempt < 2; attempt++ {
		var raw map[string]json.RawMessage
		if err := fetch(ctx, callSongSuggestions, map[string]string{
			"stationid": station.StationID,
			"k":         strconv.Itoa(limit),
		}, &raw); err != nil {
			return nil, errNotFound("no suggestions found for the given song")
		}
		if msg, hasErr := raw["error"]; hasErr {
			// error key present: upstream radio flakiness, retry once
			_ = msg
			continue
		}
		delete(raw, "stationid")

		keys := make([]string, 0, len(raw))
		for k := range raw {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		seen := map[string]bool{}
		results := make([]SongModel, 0, len(raw))
		for _, k := range keys {
			var slot struct {
				Song *SongRaw `json:"song"`
			}
			if err := json.Unmarshal(raw[k], &slot); err != nil || slot.Song == nil {
				continue
			}
			if slot.Song.ID == "" || seen[slot.Song.ID] {
				continue
			}
			seen[slot.Song.ID] = true
			results = append(results, songModel(*slot.Song))
		}
		return results, nil
	}
	return nil, errNotFound("no suggestions found for the given song")
}

// urlencodedList builds the ["<id>"] entity id list (URL-encoded by the
// query encoder) that the station endpoint expects.
func urlencodedList(id string) string {
	list := []string{id}
	b, _ := json.Marshal(list)
	return string(b)
}

// GetAlbum returns an album by numeric ID or by a JioSaavn album link.
func GetAlbum(ctx context.Context, id, link string) (AlbumModel, error) {
	var data AlbumRaw
	var err error
	switch {
	case strings.TrimSpace(link) != "":
		token, terr := lastSegment(link)
		if terr != nil {
			return AlbumModel{}, terr
		}
		err = fetch(ctx, callWebAPI, map[string]string{
			"token": token, "type": "album",
		}, &data)
	case strings.TrimSpace(id) != "":
		err = fetch(ctx, callAlbumDetails, map[string]string{"albumid": id}, &data)
	default:
		return AlbumModel{}, errBadInput("either album id or link is required")
	}

	if err != nil || (data.ID == "" && data.Title == "") {
		return AlbumModel{}, errNotFound("album not found")
	}
	return albumModel(data), nil
}

// GetPlaylist returns a playlist by numeric ID or by link, with
// paginated songs.
func GetPlaylist(ctx context.Context, id, link string, page, limit int) (PlaylistModel, error) {
	paging := map[string]string{
		"p": strconv.Itoa(nonNegative(page)),
		"n": strconv.Itoa(clampLimit(limit, searchDefaultLimit, searchMaxLimit)),
	}

	var data PlaylistRaw
	var err error
	switch {
	case strings.TrimSpace(link) != "":
		token, terr := lastSegment(link)
		if terr != nil {
			return PlaylistModel{}, terr
		}
		m := map[string]string{"token": token, "type": "playlist"}
		for k, v := range paging {
			m[k] = v
		}
		err = fetch(ctx, callWebAPI, m, &data)
	case strings.TrimSpace(id) != "":
		m := map[string]string{"listid": id}
		for k, v := range paging {
			m[k] = v
		}
		err = fetch(ctx, callPlaylistDetails, m, &data)
	default:
		return PlaylistModel{}, errBadInput("either playlist id or link is required")
	}
	if err != nil || (data.ID == "" && data.Title == "") {
		return PlaylistModel{}, errNotFound("playlist not found")
	}

	playlist := playlistModel(data)
	// Match the original behavior: songCount reflects the actually
	// returned (paginated) songs, not the full list size.
	playlist.SongCount = intPtr(StringOrNumber(strconv.Itoa(len(playlist.Songs))))
	return playlist, nil
}

func GetArtist(ctx context.Context, id, link string, page, songCount, albumCount int, sortBy, sortOrder string) (ArtistModel, error) {
	if sortBy != "" && sortBy != "popularity" && sortBy != "latest" && sortBy != "alphabetical" {
		return ArtistModel{}, errBadInput("invalid sortBy: use popularity, latest or alphabetical")
	}
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		return ArtistModel{}, errBadInput("invalid sortOrder: use asc or desc")
	}

	params := map[string]string{
		"n_song":     strconv.Itoa(clampLimit(songCount, 10, 100)),
		"n_album":    strconv.Itoa(clampLimit(albumCount, 10, 100)),
		"page":       strconv.Itoa(nonNegative(page)),
		"sort_order": sortOrDefault(sortOrder, "asc"),
		"category":   sortOrDefault(sortBy, "popularity"),
	}

	var data ArtistDetailsResponse
	var err error
	switch {
	case strings.TrimSpace(link) != "":
		token, terr := lastSegment(link)
		if terr != nil {
			return ArtistModel{}, terr
		}
		// webapi.get + type=artist only: extra paging params make
		// upstream return an empty array.
		err = fetch(ctx, callWebAPI, map[string]string{
			"token": token, "type": "artist",
		}, &data)
	case strings.TrimSpace(id) != "":
		params["artistId"] = id
		err = fetch(ctx, callArtistDetails, params, &data)
	default:
		return ArtistModel{}, errBadInput("either artist id or link is required")
	}

	if err != nil || (data.ID == "" && data.ArtistID == "") {
		return ArtistModel{}, errNotFound("artist not found")
	}
	return artistModel(data), nil
}

func GetArtistSongs(ctx context.Context, artistID string, page int, sortBy, sortOrder string) (SearchSongsResponse, error) {
	if strings.TrimSpace(artistID) == "" {
		return SearchSongsResponse{}, errBadInput("artist id is required")
	}

	var data ArtistSongsResponse
	if err := fetch(ctx, callArtistSongs, map[string]string{
		"artistId":   artistID,
		"page":       strconv.Itoa(nonNegative(page)),
		"sort_order": sortOrDefault(sortOrder, "desc"),
		"category":   sortOrDefault(sortBy, "popularity"),
	}, &data); err != nil {
		return SearchSongsResponse{}, errNotFound("artist songs not found")
	}

	seen := map[string]bool{}
	results := make([]SongModel, 0, len(data.TopSongs.Songs))
	for _, s := range data.TopSongs.Songs {
		if s.ID == "" || seen[s.ID] {
			continue
		}
		seen[s.ID] = true
		results = append(results, songModel(s))
	}
	return SearchSongsResponse{Total: data.TopSongs.Total, Start: 0, Results: results}, nil
}

func GetArtistAlbums(ctx context.Context, artistID string, page int, sortBy, sortOrder string) (SearchAlbumsResponse, error) {
	if strings.TrimSpace(artistID) == "" {
		return SearchAlbumsResponse{}, errBadInput("artist id is required")
	}

	var data ArtistAlbumsResponse
	if err := fetch(ctx, callArtistAlbums, map[string]string{
		"artistId":   artistID,
		"page":       strconv.Itoa(nonNegative(page)),
		"sort_order": sortOrDefault(sortOrder, "desc"),
		"category":   sortOrDefault(sortBy, "popularity"),
	}, &data); err != nil {
		return SearchAlbumsResponse{}, errNotFound("artist albums not found")
	}

	return SearchAlbumsResponse{
		Total:   data.TopAlbums.Total,
		Start:   0,
		Results: albumModels(data.TopAlbums.Albums),
	}, nil
}

// Trending returns the currently trending entities for a type +
// language. Fixes the "every trending request returns the same data"
// bug: type and language are forwarded as entity_type/entity_language
// and responses are cached per full query string.
func Trending(ctx context.Context, kind, language string) (TrendingResponse, error) {
	switch kind {
	case "":
		kind = "album"
	case "album", "song", "playlist":
	default:
		return TrendingResponse{}, errBadInput("invalid type: use album, song or playlist")
	}
	if strings.TrimSpace(language) == "" {
		language = "hindi"
	}

	var data TrendingListResponse
	if err := fetch(ctx, callTrending, map[string]string{
		"entity_type":     kind,
		"entity_language": language,
	}, &data); err != nil {
		return TrendingResponse{}, errNotFound("trending not found")
	}

	results := make([]TrendingItem, 0, len(data))
	for _, raw := range data {
		results = append(results, trendingItemModel(raw))
	}
	return TrendingResponse{Results: results}, nil
}
