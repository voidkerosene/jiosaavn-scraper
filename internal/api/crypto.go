package api

import (
	"crypto/des"
	"encoding/base64"
	"regexp"
	"strings"
)

// desKey is the well-known community key JioSaavn uses for its
// DES-ECB encrypted media URLs.
const desKey = "38346591"

// sizeMarker matches the size marker inside cover-art URLs.
var sizeMarkerRe = regexp.MustCompile(`150x150|50x50`)

var downloadQualities = []struct {
	bitrate string
	id      string
}{
	{"12kbps", "_12"},
	{"48kbps", "_48"},
	{"96kbps", "_96"},
	{"160kbps", "_160"},
	{"320kbps", "_320"},
}

// decryptMedia decodes a base64 DES-ECB ciphertext into a plain URL,
// stripping PKCS5 padding.
func decryptMedia(encoded string) string {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	block, err := des.NewCipher([]byte(desKey))
	if err != nil {
		return ""
	}

	decrypted := make([]byte, len(raw))
	for i := 0; i+block.BlockSize() <= len(raw); i += block.BlockSize() {
		block.Decrypt(decrypted[i:i+block.BlockSize()], raw[i:i+block.BlockSize()])
	}
	if len(decrypted) == 0 {
		return ""
	}

	// Strip PKCS5 padding, falling back to trimming trailing junk.
	n := int(decrypted[len(decrypted)-1])
	if n > 0 && n <= block.BlockSize() {
		valid := true
		for _, b := range decrypted[len(decrypted)-n:] {
			if b != byte(n) {
				valid = false
				break
			}
		}
		if valid {
			decrypted = decrypted[:len(decrypted)-n]
		}
	}
	return strings.TrimRight(string(decrypted), "\x00 \t\r\n\x01\x02\x03\x04\x05\x06\x07\x08")
}

// CreateDownloadLinks expands an encrypted media URL into one playable
// link per available bitrate.
func CreateDownloadLinks(encryptedMediaUrl string) []DownloadLink {
	if encryptedMediaUrl == "" {
		return []DownloadLink{}
	}
	base := decryptMedia(encryptedMediaUrl)
	if base == "" {
		return []DownloadLink{}
	}

	links := make([]DownloadLink, 0, len(downloadQualities))
	for _, q := range downloadQualities {
		links = append(links, DownloadLink{
			Quality: q.bitrate,
			URL:     strings.Replace(base, "_96", q.id, 1),
		})
	}
	return links
}

// CreateImageLinks rewrites a cover-art URL with the size markers
// JioSaavn CDNs accept (50x50, 150x150, 500x500), forcing https.
func CreateImageLinks(link string) []DownloadLink {
	if link == "" {
		return []DownloadLink{}
	}

	base := sanitize(link)
	var re = sizeMarkerRe

	links := make([]DownloadLink, 0, 3)
	for _, quality := range []string{"50x50", "150x150", "500x500"} {
		links = append(links, DownloadLink{
			Quality: quality,
			URL:     re.ReplaceAllString(base, quality),
		})
	}
	return links
}
