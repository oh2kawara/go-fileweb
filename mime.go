package main

import (
	"mime"
	"path/filepath"
	"strings"
)

// その他のmimeType
var otherMimeType = "application/octed-stream"

// mime
var mimeType map[string]string

func init() {
	mimeType = map[string]string{
		".html": "text/html",
		".htm":  "text/html",
		".js":   "text/javascript",
		".css":  "text/css",
		".jpeg": "image/jpeg",
		".jpg":  "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
	}
}

/**
 * 拡張氏からmime-typeを取得
 */
func GetMimeTypeByFilename(fname string) string {
	ext := strings.ToLower(filepath.Ext(fname))
	mime := mime.TypeByExtension(ext)
	if mime == "" {
		mime = mimeType[ext]
		if mime == "" {
			mime = otherMimeType
		}
	}
	return mime
}
