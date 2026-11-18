package storage

import (
	"errors"
	"io"
	"net/http"
)

// MaxImageBytes is the maximum accepted image upload size.
const MaxImageBytes int64 = 5 << 20 // 5 MiB

// ErrUnsupportedImage is returned for disallowed image types.
var ErrUnsupportedImage = errors.New("tipo de imagem não suportado (use jpeg, png, webp ou gif)")

var allowedImageExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// DetectImage sniffs the content type of an image stream, validates it against
// the allow-list and returns the MIME type and file extension. The reader is
// rewound to the start so the caller can persist it.
func DetectImage(rs io.ReadSeeker) (contentType, ext string, err error) {
	head := make([]byte, 512)
	n, err := rs.Read(head)
	if err != nil && err != io.EOF {
		return "", "", err
	}
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}
	contentType = http.DetectContentType(head[:n])
	ext, ok := allowedImageExt[contentType]
	if !ok {
		return "", "", ErrUnsupportedImage
	}
	return contentType, ext, nil
}
