// JioSaavn Scrape API: an unofficial, fast Go implementation of the
// JioSaavn music endpoint surface (search, songs, albums, playlists,
// artists, trending) with JSON metadata and audio stream URLs.
package main

import (
	"flag"
	"log"

	"jiosaavn-scraper/internal/server"
)

func main() {
	port := flag.String("port", "", "port override (wins over the PORT environment variable)")
	flag.Parse()

	cfg := server.LoadConfig()
	if *port != "" {
		cfg.Addr = ":" + *port
	}

	if err := server.Serve(cfg); err != nil {
		log.Fatal(err)
	}
}
