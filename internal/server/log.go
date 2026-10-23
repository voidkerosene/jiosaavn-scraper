package server

import (
	"log"
	"net/http"
	"time"
)

func logRequest(r *http.Request, status, size int, state string, took time.Duration) {
	log.Printf("%s %s -> %d (%d bytes, %s, took %s)",
		r.Method, r.URL.RequestURI(), status, size, state, took)
}
