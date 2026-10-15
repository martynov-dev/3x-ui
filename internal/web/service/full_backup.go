package service

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/config"
	"github.com/mhsanaei/3x-ui/v3/internal/database"
)

// FullBackupSettings are separate from upstream settings to keep fork changes isolated.
type FullBackupSettings struct {
	Enabled       bool     `json:"enabled"`
	IntervalHours int      `json:"intervalHours"`
	Keep          int      `json:"keep"`
	Paths         []string `json:"paths"`
	RemoteHost    string   `json:"remoteHost"`
	RemoteUser    string   `json:"remoteUser"`
	RemotePort    int      `json:"remotePort"`
	RemotePath    string   `json:"remotePath"`
	IdentityFile  string   `json:"identityFile"`
}

type FullBackupStatus struct {
	LastAttempt time.Time `json:"lastAttempt"`
	LastSuccess time.Time `json:"lastSuccess"`
	File        string    `json:"file"`
	Error       string    `json:"error"`
}

type FullBackupFile struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

var fullBackupMu sync.Mutex
var safeSSHHost = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*$`)
var safeSSHUser = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)
var safeSSHPath = regexp.MustCompile(`^/[a-zA-Z0-9_./-]+$`)

func fullBackupDir() string { return filepath.Join(config.GetDBFolderPath(), "full-backups") }

func (s *SettingService) GetFullBackupSettings() (FullBackupSettings, error) {
	value := FullBackupSettings{IntervalHours: 24, Keep: 7, RemotePort: 22,
		Paths: []string{"/etc/nginx", "/var/www", "/usr/share/nginx/html", "/etc/letsencrypt", "/root/cert", "/root/.acme.sh"}}
	row, err := s.getSetting("fork.fullBackup")
	if database.IsNotFound(err) {
		return value, nil
	}
	if err != nil {
		return value, err
	}
	err = json.Unmarshal([]byte(row.Value), &value)
	return value, err
}

func validateFullBackupSettings(v FullBackupSettings) error {
	if v.IntervalHours < 1 || v.IntervalHours > 8760 || v.Keep < 1 || v.Keep > 365 {
		return fmt.Errorf("intervalHours must be 1–8760; keep must be 1–365")
	}
	for _, path := range v.Paths {
		if !filepath.IsAbs(path) || filepath.Clean(path) == string(filepath.Separator) {
			return fmt.Errorf("backup source must be an absolute directory other than root: %q", path)
		}
	}
	if v.RemoteHost != "" {
		if !safeSSHHost.MatchString(v.RemoteHost) || !safeSSHUser.MatchString(v.RemoteUser) ||
			!safeSSHPath.MatchString(v.RemotePath) || strings.Contains(v.RemotePath, "..") ||
			v.RemotePort < 1 || v.RemotePort > 65535 || !filepath.IsAbs(v.IdentityFile) {
			return fmt.Errorf("invalid SSH destination: use hostname/IPv4, user, port, absolute remote directory and private key path")
		}
	}
	return nil
}

func (s *SettingService) SaveFullBackupSettings(value FullBackupSettings) error {
	if err := validateFullBackupSettings(value); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.saveSetting("fork.fullBackup", string(data))
}

func (s *SettingService) GetFullBackupStatus() (FullBackupStatus, error) {
	var value FullBackupStatus
	row, err := s.getSetting("fork.fullBackupStatus")
	if database.IsNotFound(err) {
		return value, nil
	}
	if err != nil {
		return value, err
	}
	err = json.Unmarshal([]byte(row.Value), &value)
	return value, err
}

func (s *SettingService) saveFullBackupStatus(v FullBackupStatus) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.saveSetting("fork.fullBackupStatus", string(data))
}

// RunFullBackup serializes manual and scheduled backups. A failed upload retains the local archive.
func (s *ServerService) RunFullBackup(ctx context.Context, scheduled bool) (status FullBackupStatus, err error) {
	if !fullBackupMu.TryLock() {
		if scheduled {
			return status, nil
		}
		return status, fmt.Errorf("a full backup is already running")
	}
	defer fullBackupMu.Unlock()
	settings := &SettingService{}
	v, err := settings.GetFullBackupSettings()
	if err != nil {
		return status, err
	}
	if err = validateFullBackupSettings(v); err != nil {
		return status, err
	}
	status, err = settings.GetFullBackupStatus()
	if err != nil {
		return status, err
	}
	if scheduled && (!v.Enabled || time.Since(status.LastAttempt) < time.Duration(v.IntervalHours)*time.Hour) {
		return status, nil
	}
	status.LastAttempt, status.Error = time.Now(), ""
	if err = settings.saveFullBackupStatus(status); err != nil {
		return status, err
	}
	defer func() {
		if err != nil {
			status.Error = err.Error()
		}
		if saveErr := settings.saveFullBackupStatus(status); saveErr != nil {
			err = fmt.Errorf("backup result: %v; persist status: %w", err, saveErr)
		}
	}()
	if err = os.MkdirAll(fullBackupDir(), 0o700); err != nil {
		return status, err
	}
	if err = os.Chmod(fullBackupDir(), 0o700); err != nil {
		return status, err
	}
	staging, err := os.MkdirTemp(fullBackupDir(), ".staging-")
	if err != nil {
		return status, err
	}
	defer os.RemoveAll(staging)
	db, err := s.GetDb()
	if err != nil {
		return status, err
	}
	dbName := "x-ui.db"
	if database.IsPostgres() {
		dbName = "x-ui.dump"
	}
	if err = os.WriteFile(filepath.Join(staging, dbName), db, 0o600); err != nil {
		return status, err
	}
	name := "full-" + time.Now().UTC().Format("2006-01-02_150405.000000000") + ".tar.gz"
	file, err := os.OpenFile(filepath.Join(fullBackupDir(), name+".partial"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return status, err
	}
	defer os.Remove(file.Name())
	paths := append([]string{config.GetDBFolderPath(), config.GetBinFolderPath(), v.IdentityFile}, v.Paths...)
	if exe, exeErr := os.Executable(); exeErr == nil {
		paths = append(paths, exe)
	} else {
		_ = file.Close()
		return status, exeErr
	}
	paths = append(paths, config.GetEnvFilePaths()...)
	// Panel certificates may be outside standard nginx/acme directories.
	all, settingsErr := settings.GetAllSetting()
	if settingsErr != nil {
		_ = file.Close()
		return status, settingsErr
	}
	paths = append(paths, all.WebCertFile, all.WebKeyFile, all.SubCertFile, all.SubKeyFile,
		"/etc/systemd/system/x-ui.service", "/usr/lib/systemd/system/x-ui.service", "/etc/init.d/x-ui", "/usr/bin/x-ui")
	nginxPaths, discoverErr := discoverNginxBackupPaths(ctx)
	if discoverErr != nil {
		_ = file.Close()
		return status, discoverErr
	}
	paths = append(paths, nginxPaths...)
	site, siteErr := settings.GetNginxSitePaths()
	if siteErr != nil {
		_ = file.Close()
		return status, siteErr
	}
	paths = append(paths, site.Root, site.ConfigPath)
	err = writeFullBackup(ctx, file, staging, dbName, paths)
	closeErr := file.Close()
	if err != nil {
		return status, err
	}
	if closeErr != nil {
		return status, closeErr
	}
	if err = os.Rename(file.Name(), filepath.Join(fullBackupDir(), name)); err != nil {
		return status, err
	}
	status.File = name
	if err = pruneFullBackups(v.Keep); err != nil {
		return status, err
	}
	if v.RemoteHost != "" {
		if err = uploadFullBackup(ctx, v, name); err != nil {
			return status, err
		}
	}
	status.LastSuccess = time.Now()
	return status, nil
}

// writeFullBackup uses a database snapshot, never the live SQLite/WAL files.
// Symlinks and their resolved targets are archived, including external certificate paths.
func writeFullBackup(ctx context.Context, out io.Writer, staging, dbName string, paths []string) error {
	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	seen := map[string]bool{}
	seenRoots := map[string]bool{}
	missing := []string{}
	backupDir, err := filepath.Abs(fullBackupDir())
	if err != nil {
		return err
	}
	liveDB, err := filepath.Abs(config.GetDBPath())
	if err != nil {
		return err
	}
	add := func(path, name string, info fs.FileInfo) error {
		if seen[name] {
			return nil
		}
		seen[name] = true
		link := ""
		if info.Mode()&os.ModeSymlink != 0 {
			var err error
			link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		header.Name = name
		if err = tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tw, f)
		closeErr := f.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}
	dbPath := filepath.Join(staging, dbName)
	info, err := os.Stat(dbPath)
	if err != nil {
		return err
	}
	if err = add(dbPath, "database/"+dbName, info); err != nil {
		return err
	}
	for i := 0; i < len(paths); i++ {
		root := paths[i]
		if root == "" {
			continue
		}
		root, err = filepath.Abs(root)
		if err != nil {
			return err
		}
		if seenRoots[root] {
			continue
		}
		seenRoots[root] = true
		if root == filepath.VolumeName(root)+string(filepath.Separator) {
			return fmt.Errorf("refusing to archive the filesystem root")
		}
		if _, statErr := os.Lstat(root); os.IsNotExist(statErr) {
			missing = append(missing, root)
			continue
		} else if statErr != nil {
			return statErr
		}
		err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if path == backupDir || path == staging || entry.IsDir() && (entry.Name() == "logs" || entry.Name() == ".git" || strings.HasPrefix(entry.Name(), ".x-ui-backup-")) {
				return filepath.SkipDir
			}
			if path == liveDB || path == liveDB+"-wal" || path == liveDB+"-shm" {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() && !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
				return nil
			}
			if info.Mode()&os.ModeSymlink != 0 {
				target, err := filepath.EvalSymlinks(path)
				if os.IsNotExist(err) {
					missing = append(missing, path)
					return nil
				}
				if err != nil {
					return err
				}
				paths = append(paths, target)
			}
			name := "files/" + strings.TrimLeft(filepath.ToSlash(path), "/")
			if filepath.VolumeName(path) != "" {
				name = "files/" + strings.ReplaceAll(strings.TrimLeft(filepath.ToSlash(path), "/"), ":", "")
			}
			return add(path, name, info)
		})
		if err != nil {
			return fmt.Errorf("archive %s: %w", root, err)
		}
	}
	manifest, err := json.MarshalIndent(struct {
		Version string    `json:"version"`
		Created time.Time `json:"created"`
		Sources []string  `json:"sources"`
		Missing []string  `json:"missing"`
	}{config.GetPanelVersion(), time.Now().UTC(), paths, missing}, "", "  ")
	if err != nil {
		return err
	}
	if err = tw.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0o600, Size: int64(len(manifest))}); err != nil {
		return err
	}
	if _, err = tw.Write(manifest); err != nil {
		return err
	}
	if err = tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

func ListFullBackups() ([]FullBackupFile, error) {
	files := []FullBackupFile{}
	entries, err := os.ReadDir(fullBackupDir())
	if os.IsNotExist(err) {
		return files, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !isFullBackupName(entry.Name()) || !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		files = append(files, FullBackupFile{entry.Name(), info.Size(), info.ModTime()})
	}
	slices.SortFunc(files, func(a, b FullBackupFile) int { return strings.Compare(b.Name, a.Name) })
	return files, nil
}

var fullBackupNamePattern = regexp.MustCompile(`^full-\d{4}-\d{2}-\d{2}_\d{6}\.\d{9}\.tar\.gz$`)

func isFullBackupName(name string) bool { return fullBackupNamePattern.MatchString(name) }

func FullBackupPath(name string) (string, error) {
	if !isFullBackupName(name) {
		return "", fmt.Errorf("invalid backup filename")
	}
	path := filepath.Join(fullBackupDir(), name)
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("backup is not a regular file")
	}
	return path, nil
}

func pruneFullBackups(keep int) error {
	files, err := ListFullBackups()
	if err != nil {
		return err
	}
	for _, file := range files[min(keep, len(files)):] {
		if err := os.Remove(filepath.Join(fullBackupDir(), file.Name)); err != nil {
			return err
		}
	}
	return nil
}

func uploadFullBackup(ctx context.Context, v FullBackupSettings, name string) error {
	// Strict host verification is mandatory. Enroll the host key on the server first.
	args := []string{"-B", "-P", fmt.Sprint(v.RemotePort), "-i", v.IdentityFile,
		"-o", "StrictHostKeyChecking=yes", "-o", "ConnectTimeout=20", "--", filepath.Join(fullBackupDir(), name),
		v.RemoteUser + "@" + v.RemoteHost + ":" + strings.TrimRight(v.RemotePath, "/") + "/" + name + ".partial"}
	output, err := exec.CommandContext(ctx, "scp", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("SSH upload failed (local archive retained): %w: %s", err, output)
	}
	remoteFile := strings.TrimRight(v.RemotePath, "/") + "/" + name
	args = []string{"-p", fmt.Sprint(v.RemotePort), "-i", v.IdentityFile, "-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=yes", "-o", "ConnectTimeout=20",
		v.RemoteUser + "@" + v.RemoteHost, "mv -- " + remoteFile + ".partial " + remoteFile}
	output, err = exec.CommandContext(ctx, "ssh", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("finalize remote archive (local archive retained): %w: %s", err, output)
	}
	return nil
}
