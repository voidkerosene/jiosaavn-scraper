# JioSaavn Scrape API

An unofficial, fast API exposing the JioSaavn music catalog: entity search,
song, album, playlist and artist details, per-language discovery feeds
(stations, channels, charts, new releases, top playlists), trending feeds, and
song metadata with audio stream URLs at every available quality (12 to 320
kbps). Media URLs are provided through the `downloadUrl` field, matching the
JSON surface of the original project it clones. Synced and plain lyrics are
served through an integration with LRCLIB (https://lrclib.net).

The implementation is written in Go and depends only on the Go standard
library.

## Repository layout

```
main.go                       entrypoint
internal/api/                 upstream JioSaavn client, crypto, transforms,
                              endpoint services, per-language browse feeds
internal/lrclib/              LRCLIB lyrics client (public free API)
internal/server/              configuration, cache with singleflight,
                              request pipeline, handlers, JSON documentation
.env.example                  configuration reference, copy to .env
Makefile                      build, test, format, vet, run targets
.github/workflows/ci.yml      continuous integration (build, vet, test)
```

## Requirements

Go 1.27 or newer.

## Running

```sh
cp .env.example .env   # optional; defaults apply without it
go run .               # listens on the port configured by PORT, default 3000
```

The API documentation is served as JSON at the root path (`/`).

Configuration variables (all optional, each documented in `.env.example`):

| Variable           | Default | Description                          |
| ------------------ | ------- | ------------------------------------ |
| `PORT`             | `3000`  | HTTP listen port                     |
| `LOG_FILE`         | unset   | additional log file; stdout always  |
| `CACHE_TTL_MINUTES`| `5`     | response cache lifetime              |
| `CACHE_MAX_SIZE`   | `5000`  | cached response cap before eviction  |

A `-port` flag is also accepted and takes precedence over `PORT`.

## Endpoints

All endpoints live under `/api` and respond with
`{ "success": true, "data": ... }`. Errors return
`{ "success": false, "message": "..." }` with a matching HTTP status code.
Every request is reported in the configured log sinks with method, path,
status, payload size, cache state and duration.

Languages accepted wherever a `language` parameter appears include: hindi,
english, punjabi, tamil, telugu, malayalam, kannada, bengali, marathi, urdu,
bhojpuri, haryanvi, rajasthani, odia, assamese, gujarati, korean and arabic.

### Discover

| Endpoint             | Params                                                                                    |
| -------------------- | ----------------------------------------------------------------------------------------- |
| `GET /api/browse`    | `section` (`all`, `stations`, `channels`, `charts`, `newAlbums`, `newTrending`, `topPlaylists`, `topShows`, default `all`), `language` |
| `GET /api/stations`  | `language`                                                                                |
| `GET /api/channels`  | `language`                                                                                |
| `GET /api/charts`    | `language`                                                                                |
| `GET /api/new-releases` | `language`                                                                             |
| `GET /api/new-trending` | `language`                                                                             |
| `GET /api/top-playlists` | `language`, `page`, `limit` (max 250)                                                 |
| `GET /api/top-shows` | `language`                                                                                |
| `GET /api/trending`  | `type` (`album`, `song`, `playlist`), `language`                                          |

Notes:

- `/api/browse` and the section aliases mirror the JioSaavn home page for the
  requested language. The `section` parameter accepts
  `stations`, `channels`, `charts`, `newAlbums`, `newTrending`,
  `topPlaylists`, `topShows`, or `all`.
- Chart entries are playlists. Fetching a chart through
  `/api/playlists?id=<chartId>` returns its ranked songs.
- `/api/trending` and `/api/top-playlists` are confirmed language-aware at the
  upstream feed level. Some regions lock additional home sections to a default
  language; the `language` parameter is forwarded regardless.

### Search

| Endpoint                     | Params                                   |
| ---------------------------- | ---------------------------------------- |
| `GET /api/search`            | `query`                                  |
| `GET /api/search/songs`      | `query`, `page` (default 0), `limit` (max 100) |
| `GET /api/search/albums`     | `query`, `page`, `limit`                 |
| `GET /api/search/artists`    | `query`, `page`, `limit`                 |
| `GET /api/search/playlists`  | `query`, `page`, `limit`                 |

### Songs, albums, playlists, artists

| Endpoint                          | Params                                                    |
| --------------------------------- | --------------------------------------------------------- |
| `GET /api/songs`                  | `ids` (comma separated) or `link`                         |
| `GET /api/songs/{id}`             | none                                                      |
| `GET /api/songs/{id}/suggestions` | `limit`                                                   |
| `GET /api/albums`                 | `id` or `link`                                            |
| `GET /api/playlists`              | `id` or `link`, `page`, `limit`                           |
| `GET /api/artists`                | `id` or `link`, `page`, `songCount`, `albumCount`, `sortBy`, `sortOrder` |
| `GET /api/artists/{id}`           | `songCount`, `albumCount`                                 |
| `GET /api/artists/{id}/songs`     | `sortBy`, `sortOrder`                                     |
| `GET /api/artists/{id}/albums`    | `sortBy`, `sortOrder`                                     |

### Lyrics

| Endpoint                     | Params                                            |
| ---------------------------- | ------------------------------------------------- |
| `GET /api/songs/{id}/lyrics` | none; matches the song on LRCLIB by title, artist, album and duration |
| `GET /api/lyrics`            | `query` (free text search on LRCLIB)               |

Lyric responses contain `syncedLyrics` in LRC format and `plainLyrics`.

## Documentation endpoints

`GET /` returns the complete, grouped documentation for all 25 endpoints as
JSON, including parameters, defaults and descriptions.

## Caching and performance

- Responses are cached for `CACHE_TTL_MINUTES` minutes, keyed by the full
  method and request URI, so different parameter combinations never collide.
- Concurrent identical requests share a single upstream fetch (singleflight);
  the `X-Cache` response header reports `hit`, `shared` or `miss`. Errors are
  never cached and report `bypass`.
- Responses larger than 256 bytes are served gzip encoded when the client
  requests it.
- The upstream HTTP client uses connection pooling with keep-alive.

## Changes from the original project

| Original behavior                                     | This implementation                                                   |
| ----------------------------------------------------- | --------------------------------------------------------------------- |
| Search returned mismatched entity types, duplicates and HTML artifacts | Results are filtered to the requested entity type, deduplicated and unescaped |
| `/trending` returned the same payload for every type and language | `entity_type` and `entity_language` are forwarded and the cache keys on the full query string |
| `/songs/{id}/suggestions` crashed on an upstream radio error | The station flow retries once, then returns a clean 404          |
| A 320 kbps stream URL was offered even when the master did not include it | The 320 kbps entry only appears when the upstream flag confirms availability |
| A new upstream request per inbound request             | Response cache, singleflight deduplication, connection pooling, gzip   |
| Numeric fields leaked `NaN`, `Infinity` or scientific notation | Numeric fields are strictly parsed; unknown values become `null` |
| No lyrics, charts or home page feeds                   | Browse endpoints mirror the JioSaavn home page; lyrics served from LRCLIB |

## Build and quality checks

```sh
make build   # compile to bin/jiosaavn-scraper
make test    # run the test suite
make fmt     # gofmt the tree
make vet     # go vet
```

## Example

```sh
curl -s 'http://localhost:3000/api/search/songs?query=believer&limit=2' \
  | jq '.data.results[0] | {id, name, downloadUrl: .downloadUrl[-1]}'
```

```json
{
  "id": "4xIrHCBD",
  "name": "Believer",
  "downloadUrl": {
    "quality": "320kbps",
    "url": "https://aac.saavncdn.com/.../xxx_320.mp4"
  }
}
```

Lyrics with timestamps:

```sh
curl -s 'http://localhost:3000/api/songs/3IoDK8qI/lyrics' | jq -r '.data.syncedLyrics' | head -3
```

```
[00:09.52] If you wanna run away with me, I know a galaxy
[00:11.77] And I can take you for a ride
[00:14.02] I had a premonition that we fell into a rhythm
```

## Disclaimer

Unofficial and not affiliated with JioSaavn or LRCLIB. Provided for
educational and personal use. Respect the terms of service of any upstream
service used.

## License

MIT. See LICENSE.
