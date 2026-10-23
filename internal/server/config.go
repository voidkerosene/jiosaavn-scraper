package server

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is runtime configuration, read from the process environment.
// A local .env file is loaded first for convenience; real environment
// variables always take precedence. Every value has a default, so the
// binary runs with no configuration at all.
type Config struct {
	Addr         string
	LogFile      string // optional additional log destination; stdout is always written to
	CacheTTL     time.Duration
	CacheMaxSize int
}

// Supported environment variables (see .env.example):
//
//	PORT               HTTP listen port            (default 3000)
//	LOG_FILE           extra log sink file path    (default stdout only)
//	CACHE_TTL_MINUTES  response cache lifetime     (default 5)
//	CACHE_MAX_SIZE     cache entry cap             (default 5000)
func LoadConfig() Config {
	loadDotEnv(".env")
	return Config{
		Addr:         ":" + strEnv("PORT", "3000"),
		LogFile:      strEnv("LOG_FILE", ""),
		CacheTTL:     time.Duration(intEnv("CACHE_TTL_MINUTES", 5)) * time.Minute,
		CacheMaxSize: intEnv("CACHE_MAX_SIZE", 5000),
	}
}

// loadDotEnv parses a simple KEY=VALUE file (comments with #, optional
// export prefix, no nesting). It never overrides an existing variable.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // no .env file is a normal, silent case
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if k == "" {
			continue
		}
		if _, exists := os.LookupEnv(k); !exists {
			if err := os.Setenv(k, v); err != nil {
				log.Printf("config: could not set %s: %v", k, err)
			}
		}
	}
}

func strEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func intEnv(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("config: %s=%q is not a number, using default %d", key, v, def)
		return def
	}
	return n
}
