package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/config"
)

func TestFullBackupUsesSnapshotAndExcludesLiveDatabaseAndArchives(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", root)
	staging := t.TempDir()
	for path, content := range map[string]string{
		filepath.Join(staging, "x-ui.db"):            "consistent snapshot",
		config.GetDBPath():                           "live database",
		config.GetDBPath() + "-wal":                  "uncommitted WAL",
		filepath.Join(fullBackupDir(), "old.tar.gz"): "old archive",
		filepath.Join(root, "nginx", "site.conf"):    "server {}",
		filepath.Join(root, "site", "index.html"):    "website",
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	missing := filepath.Join(root, "missing")
	if err := writeFullBackup(context.Background(), &out, staging, "x-ui.db", []string{root, root, missing}); err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(&out)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	files := map[string]string{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if _, exists := files[header.Name]; exists {
			t.Fatalf("duplicate entry: %s", header.Name)
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		files[header.Name] = string(data)
		if strings.Contains(string(data), "live database") || strings.Contains(string(data), "uncommitted WAL") || strings.Contains(header.Name, "full-backups") {
			t.Fatalf("live data or recursive archive: %s", header.Name)
		}
	}
	if files["database/x-ui.db"] != "consistent snapshot" {
		t.Fatal("database snapshot missing")
	}
	foundSite, foundNginx := false, false
	for name, content := range files {
		if strings.HasSuffix(name, "/site/index.html") && content == "website" {
			foundSite = true
		}
		if strings.HasSuffix(name, "/nginx/site.conf") && content == "server {}" {
			foundNginx = true
		}
	}
	if !foundSite || !foundNginx {
		t.Fatal("website or nginx missing")
	}
	var manifest struct {
		FormatVersion int      `json:"formatVersion"`
		OS            string   `json:"os"`
		Arch          string   `json:"arch"`
		DBPath        string   `json:"dbPath"`
		Executable    string   `json:"executable"`
		Missing       []string `json:"missing"`
	}
	if err := json.Unmarshal([]byte(files["manifest.json"]), &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Missing) != 1 || manifest.Missing[0] != missing {
		t.Fatalf("missing sources: %v", manifest.Missing)
	}
	if manifest.FormatVersion != 1 || manifest.OS != runtime.GOOS || manifest.Arch != runtime.GOARCH || manifest.DBPath != filepath.Clean(config.GetDBPath()) || manifest.Executable == "" {
		t.Fatalf("restore metadata missing: %+v", manifest)
	}
}

func TestFullBackupRetentionAndDownloadPaths(t *testing.T) {
	t.Setenv("XUI_DB_FOLDER", t.TempDir())
	if err := os.MkdirAll(fullBackupDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	names := []string{"full-2026-01-01_010000.000000000.tar.gz", "full-2026-01-02_010000.000000000.tar.gz", "full-2026-01-03_010000.000000000.tar.gz"}
	for _, name := range append(names, "keep.txt", names[0]+".partial") {
		if err := os.WriteFile(filepath.Join(fullBackupDir(), name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := pruneFullBackups(2); err != nil {
		t.Fatal(err)
	}
	files, err := ListFullBackups()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Name != names[2] || files[1].Name != names[1] {
		t.Fatalf("retention result: %v", files)
	}
	if _, err := os.Stat(filepath.Join(fullBackupDir(), "keep.txt")); err != nil {
		t.Fatal("unrelated file removed", err)
	}
	for _, name := range []string{"../x-ui.db", names[0] + ".partial", "x-ui.db"} {
		if _, err := FullBackupPath(name); err == nil {
			t.Fatalf("accepted unsafe download %q", name)
		}
	}
}

func TestFullBackupRejectsUnsafeDestination(t *testing.T) {
	v := FullBackupSettings{IntervalHours: 24, Keep: 7, RemoteHost: "backup.example.com", RemoteUser: "backup", RemotePort: 22, RemotePath: "/backups", IdentityFile: filepath.Join(t.TempDir(), "key")}
	if err := validateFullBackupSettings(v); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*FullBackupSettings){
		func(v *FullBackupSettings) { v.RemoteHost = "-oProxyCommand=bad" },
		func(v *FullBackupSettings) { v.RemotePath = "/backups;touch /tmp/bad" },
		func(v *FullBackupSettings) { v.RemotePath = "/backups/../other" },
		func(v *FullBackupSettings) { v.Keep = 0 },
		func(v *FullBackupSettings) { v.IntervalHours = 0 },
	} {
		bad := v
		mutate(&bad)
		if err := validateFullBackupSettings(bad); err == nil {
			t.Fatalf("accepted invalid settings: %+v", bad)
		}
	}
}
