package api

import (
	"strings"
	"testing"
)

// Vector taken live from search.getResults (api_version=4).
const liveCiphertext = "ID2ieOjCrwfgWvL5sXl4B1ImC5QfbsDySan+n+AW12BvOaQj7cuGfg8Ed085rYUtqDj8DQY3nIMQdr42ScGdtRw7tS9a8Gtq"

func TestDecryptMedia(t *testing.T) {
	got := decryptMedia(liveCiphertext)
	want := "https://aac.saavncdn.com/450/f467e05e2825cec2203546333e0d0550_96.mp4"
	if got != want {
		t.Fatalf("decrypt: got %q, want %q", got, want)
	}
}

func TestCreateDownloadLinks(t *testing.T) {
	links := CreateDownloadLinks(liveCiphertext)
	if len(links) != 5 {
		t.Fatalf("expected 5 quality links, got %d", len(links))
	}
	if links[0].Quality != "12kbps" || links[4].Quality != "320kbps" {
		t.Fatalf("unexpected quality order: %v", links)
	}
	if !strings.HasSuffix(links[4].URL, "_320.mp4") {
		t.Fatalf("320 link should end _320.mp4, got %s", links[4].URL)
	}
	if len(CreateDownloadLinks("")) != 0 {
		t.Fatal("empty encrypted url should produce no links")
	}
}

func TestCreateImageLinks(t *testing.T) {
	src := "https://c.saavncdn.com/450/foo-150x150.jpg"
	links := CreateImageLinks(src)
	if len(links) != 3 {
		t.Fatalf("expected 3 image links, got %d", len(links))
	}
	checks := map[string]string{
		"50x50":   "https://c.saavncdn.com/450/foo-50x50.jpg",
		"150x150": "https://c.saavncdn.com/450/foo-150x150.jpg",
		"500x500": "https://c.saavncdn.com/450/foo-500x500.jpg",
	}
	for _, l := range links {
		if checks[l.Quality] != l.URL {
			t.Fatalf("quality %s: got %s, want %s", l.Quality, l.URL, checks[l.Quality])
		}
	}
}

func TestSanitize(t *testing.T) {
	cases := map[string]string{
		"Gehra Hua (From &quot;Dhurandhar&quot;)": `Gehra Hua (From "Dhurandhar")`,
		"<% junk %>": "",
		"A &amp; B":  "A & B",
		" ok ":       "ok",
	}
	for in, want := range cases {
		if got := sanitize(in); got != want {
			t.Fatalf("sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLastSegment(t *testing.T) {
	if tok, _ := lastSegment("https://www.jiosaavn.com/song/houdini/OgwhbhtDRwM"); tok != "OgwhbhtDRwM" {
		t.Fatalf("song token: %s", tok)
	}
	if tok, _ := lastSegment("https://www.jiosaavn.com/album/future-nostalgia/ITIyo-GDr7A_"); tok != "ITIyo-GDr7A_" {
		t.Fatalf("album token: %s", tok)
	}
	if _, err := lastSegment("totally invalid"); err == nil {
		t.Fatal("expected error for invalid link")
	}
}
