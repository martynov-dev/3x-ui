package middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Compression leaves archive downloads byte-for-byte intact. The gzip middleware
// can mistake the first chunk of a gzip file for an already encoded HTTP body,
// then compress subsequent chunks, corrupting the downloaded archive.
func Compression() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{
		".png", ".gif", ".jpeg", ".jpg", ".gz",
	}))
}
