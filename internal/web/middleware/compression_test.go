package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"math/rand"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	gzipmiddleware "github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func TestArchiveDownloadCompression(t *testing.T) {
	data := make([]byte, 1024*1024)
	rand.New(rand.NewSource(1)).Read(data)
	var archive bytes.Buffer
	writer := gzip.NewWriter(&archive)
	if _, err := writer.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "full.tar.gz")
	if err := os.WriteFile(path, archive.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	for _, legacy := range []bool{true, false} {
		router := gin.New()
		if legacy {
			router.Use(gzipmiddleware.Gzip(gzipmiddleware.DefaultCompression))
		} else {
			router.Use(Compression())
		}
		router.GET("/panel/api/maintenance/backups/download/full.tar.gz", func(c *gin.Context) {
			c.FileAttachment(path, "full.tar.gz")
		})
		request := httptest.NewRequest("GET", "/panel/api/maintenance/backups/download/full.tar.gz", nil)
		request.Header.Set("Accept-Encoding", "gzip")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if legacy {
			if bytes.Equal(response.Body.Bytes(), archive.Bytes()) {
				t.Fatal("legacy middleware did not reproduce corruption")
			}
			continue
		}
		if response.Code != 200 || response.Header().Get("Content-Encoding") != "" {
			t.Fatalf("unexpected response: %d %v", response.Code, response.Header())
		}
		if !bytes.Equal(response.Body.Bytes(), archive.Bytes()) {
			t.Fatal("download differs from original archive")
		}
	}
}

func TestCompressionStillCompressesJSON(t *testing.T) {
	router := gin.New()
	router.Use(Compression())
	router.GET("/api/settings", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	request := httptest.NewRequest("GET", "/api/settings", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("JSON compression disabled")
	}
	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil || string(body) != `{"ok":true}` {
		t.Fatalf("invalid JSON response: %s, %v", body, err)
	}
}
