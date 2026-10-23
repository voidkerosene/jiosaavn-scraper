package server

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"jiosaavn-scraper/internal/api"
)

// kinded handlers return (payload, error) and never touch
// ResponseWriter; the pipeline renders, caches, compresses, logs.
type kinded = func(r *http.Request) (any, error)

type okPayload struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type errorPayload struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func httpStatus(err error) (int, string) {
	var he *api.HTTPError
	if errors.As(err, &he) {
		return he.Code, he.Message
	}
	return 500, "internal server error"
}

// render marshals any payload to pretty JSON bytes.
func render(status int, v any) []byte {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		body, _ = json.Marshal(errorPayload{Success: false, Message: "internal encoding error"})
	}
	return body
}

// respond renders the success or error envelope for a handler result.
func respond(err error, payload any) (cacheEntry, error) {
	if err != nil {
		status, msg := httpStatus(err)
		return cacheEntry{
			body:   render(status, errorPayload{Success: false, Message: msg}),
			status: status,
		}, err
	}
	return cacheEntry{
		body:   render(http.StatusOK, okPayload{Success: true, Data: payload}),
		status: http.StatusOK,
	}, nil
}

const minGzipSize = 256

// wantGzip is evaluated before WriteHeader flushes the header block.
func wantGzip(r *http.Request, body []byte) bool {
	return len(body) >= minGzipSize && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

// kindedPipeline is the whole request pipeline for API endpoints:
// flight-shared cache -> handler -> JSON render -> gzip -> write -> log.
func pipeline(fn kinded, c *cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		var entry cacheEntry
		var err error
		var state string

		compute := func() (cacheEntry, error) {
			payload, perr := fn(r)
			return respond(perr, payload)
		}

		if r.Method == http.MethodGet {
			entry, state, err = c.fetch(cacheKey(r), compute)
		} else {
			entry, err = compute()
			state = "miss"
		}
		status := entry.status
		if err != nil {
			status, _ = httpStatus(err)
		}
		if status == 0 { // defensive: never write header 0
			status = http.StatusInternalServerError
		}

		body := entry.body
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("X-Cache", cacheHeader(state, status))
		gzipIt := wantGzip(r, body)
		if gzipIt {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")
		}
		w.WriteHeader(status)
		if gzipIt {
			gz := gzip.NewWriter(w)
			gz.Write(body)
			gz.Close()
		} else {
			w.Write(body)
		}

		logRequest(r, status, len(body), state, time.Since(start))
	}
}

func cacheHeader(state string, status int) string {
	if status != 200 {
		return "bypass" // errors are never cached
	}
	return state
}

// direct wraps simple handlers (docs root) with headers + log only.
func direct(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		next(w, r)
		logRequest(r, http.StatusOK, 0, "docs", time.Since(start))
	}
}
