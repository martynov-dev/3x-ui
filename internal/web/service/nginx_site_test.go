package service

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNginxPathsIncludeSitesAndCertificatesWithoutVariableRoots(t *testing.T) {
	got := nginxDirectivePaths(`server {
    root /var/www/example;
    alias "/srv/website assets/";
    ssl_certificate /etc/letsencrypt/live/example/fullchain.pem;
    ssl_certificate_key '/etc/letsencrypt/live/example/privkey.pem';
    root /srv/$host;
    root /;
}`)
	want := []string{"/var/www/example", "/srv/website assets/", "/etc/letsencrypt/live/example/fullchain.pem", "/etc/letsencrypt/live/example/privkey.pem"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
}

func TestRestoreSiteFilesRevertsReplacedAndNewFiles(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "site.conf")
	newFile := filepath.Join(dir, "index.html")
	if err := os.WriteFile(old, []byte("old config"), 0o600); err != nil {
		t.Fatal(err)
	}
	original, err := rememberSiteFile(old)
	if err != nil {
		t.Fatal(err)
	}
	absent, err := rememberSiteFile(newFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeSiteFile(old, []byte("bad config"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeSiteFile(newFile, []byte("new site"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := restoreSiteFiles([]siteOriginal{original, absent}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(old)
	if err != nil || string(data) != "old config" {
		t.Fatalf("restore: %q, %v", data, err)
	}
	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Fatalf("new file not removed: %v", err)
	}
}

func TestSitePathContainment(t *testing.T) {
	root := filepath.Join(t.TempDir(), "www")
	if !withinPath(root, filepath.Join(root, "site", "index.html")) {
		t.Fatal("rejected site path")
	}
	for _, path := range []string{root + "-other", filepath.Join(root, "..", "secret")} {
		if withinPath(root, path) {
			t.Fatalf("accepted escape %s", path)
		}
	}
}
