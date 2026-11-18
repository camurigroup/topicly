package storage

import (
	"net/url"
	"strings"
)

// FilesPrefix is the path under which the local provider serves uploads.
const FilesPrefix = "/files/"

// RelativeFileURL converts an absolute URL pointing at the local file server
// ("http(s)://any-host/files/…") into the host-independent "/files/…" form.
// Stored URLs used to embed the API host and port at upload time, which broke
// every image when the host or port changed. Other URLs (MinIO/CDN, already
// relative, or suspicious paths) are returned unchanged.
func RelativeFileURL(raw string) (string, bool) {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return raw, false
	}
	u, err := url.Parse(raw)
	if err != nil || !strings.HasPrefix(u.Path, FilesPrefix) || strings.Contains(u.Path, "..") {
		return raw, false
	}
	return u.Path, true
}
