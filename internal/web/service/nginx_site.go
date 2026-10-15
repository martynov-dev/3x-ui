package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
)

type NginxSite struct {
	ConfigPath string `json:"configPath"`
	Root       string `json:"root"`
	Config     string `json:"config"`
	HTML       string `json:"html"`
}

var nginxSiteMu sync.Mutex

var nginxFileMarker = regexp.MustCompile(`(?m)^# configuration file (.+):$`)
var nginxPathDirective = regexp.MustCompile(`(?m)^\s*(?:root|alias|ssl_certificate|ssl_certificate_key|ssl_trusted_certificate)\s+(?:"([^"]+)"|'([^']+)'|([^;\s]+))\s*;`)
var nginxDocumentRootDirective = regexp.MustCompile(`(?m)^\s*(?:root|alias)\s+[^;]+;`)

func nginxDump(ctx context.Context) (string, error) {
	if _, err := exec.LookPath("nginx"); err != nil {
		return "", nil
	}
	output, err := exec.CommandContext(ctx, "nginx", "-T").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("discover nginx configuration: %w: %s", err, output)
	}
	return string(output), nil
}

func nginxDirectivePaths(dump string) []string {
	paths := []string{}
	for _, match := range nginxPathDirective.FindAllStringSubmatch(dump, -1) {
		for _, value := range match[1:] {
			if path.IsAbs(value) && path.Clean(value) != "/" && !strings.Contains(value, "$") {
				paths = append(paths, value)
			}
		}
	}
	return paths
}

func discoverNginxBackupPaths(ctx context.Context) ([]string, error) {
	dump, err := nginxDump(ctx)
	if err != nil {
		return nil, err
	}
	paths := nginxDirectivePaths(dump)
	for _, match := range nginxFileMarker.FindAllStringSubmatch(dump, -1) {
		if filepath.IsAbs(match[1]) {
			paths = append(paths, match[1])
		}
	}
	return paths, nil
}

// DiscoverNginxSites lists enabled server blocks without choosing one to overwrite.
func DiscoverNginxSites(ctx context.Context) ([]NginxSite, error) {
	dump, err := nginxDump(ctx)
	if err != nil {
		return nil, err
	}
	sites := []NginxSite{}
	markers := nginxFileMarker.FindAllStringSubmatchIndex(dump, -1)
	for i, marker := range markers {
		path := dump[marker[2]:marker[3]]
		end := len(dump)
		if i+1 < len(markers) {
			end = markers[i+1][0]
		}
		block := dump[marker[1]:end]
		if !strings.Contains(block, "server {") && !strings.Contains(block, "server{") {
			continue
		}
		if !withinPath("/etc/nginx", path) || path == "/etc/nginx/nginx.conf" {
			continue
		}
		root := "/var/www/3x-ui-site"
		for _, value := range nginxDirectivePaths(strings.Join(nginxDocumentRootDirective.FindAllString(block, -1), "\n")) {
			if withinPath("/var/www", value) || withinPath("/usr/share/nginx/html", value) {
				root = value
				break
			}
		}
		sites = append(sites, NginxSite{ConfigPath: path, Root: root})
	}
	return sites, nil
}

func (s *SettingService) GetNginxSitePaths() (NginxSite, error) {
	v := NginxSite{ConfigPath: "/etc/nginx/conf.d/3x-ui-site.conf", Root: "/var/www/3x-ui-site"}
	row, err := s.getSetting("fork.nginxSite")
	if database.IsNotFound(err) {
		return v, nil
	}
	if err != nil {
		return v, err
	}
	err = json.Unmarshal([]byte(row.Value), &v)
	return v, err
}

func withinPath(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// Resolve existing symlinks, including parent directories of newly created files.
func resolveSitePath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("path must be absolute")
	}
	path = filepath.Clean(path)
	current := path
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			rel, err := filepath.Rel(current, path)
			if err != nil {
				return "", err
			}
			return filepath.Join(resolved, rel), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", err
		}
		current = parent
	}
}

func validateSitePaths(v NginxSite) (string, string, error) {
	configPath, err := resolveSitePath(v.ConfigPath)
	if err != nil {
		return "", "", err
	}
	if !withinPath("/etc/nginx", configPath) || configPath == "/etc/nginx/nginx.conf" {
		return "", "", fmt.Errorf("select a site configuration inside /etc/nginx; nginx.conf is not editable here")
	}
	index, err := resolveSitePath(filepath.Join(v.Root, "index.html"))
	if err != nil {
		return "", "", err
	}
	if !withinPath("/var/www", index) && !withinPath("/usr/share/nginx/html", index) {
		return "", "", fmt.Errorf("site root must be inside /var/www or /usr/share/nginx/html")
	}
	if configPath == "/etc/nginx" {
		return "", "", fmt.Errorf("select a configuration file")
	}
	return configPath, index, nil
}

func (s *SettingService) ReadNginxSite(v NginxSite) (NginxSite, error) {
	nginxSiteMu.Lock()
	defer nginxSiteMu.Unlock()
	configPath, index, err := validateSitePaths(v)
	if err != nil {
		return v, err
	}
	read := func(path string) (string, error) {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return "", nil
		}
		return string(data), err
	}
	if v.Config, err = read(configPath); err != nil {
		return v, err
	}
	if v.HTML, err = read(index); err != nil {
		return v, err
	}
	return v, nil
}

type siteOriginal struct {
	path    string
	data    []byte
	mode    os.FileMode
	existed bool
}

func rememberSiteFile(path string) (siteOriginal, error) {
	v := siteOriginal{path: path, mode: 0o644}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return v, nil
	}
	if err != nil {
		return v, err
	}
	if !info.Mode().IsRegular() {
		return v, fmt.Errorf("not a regular file: %s", path)
	}
	v.existed, v.mode = true, info.Mode().Perm()
	v.data, err = os.ReadFile(path)
	return v, err
}

func writeSiteFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".3x-ui-site-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(mode); err != nil {
		_ = f.Close()
		return err
	}
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func restoreSiteFiles(originals []siteOriginal) error {
	for _, original := range originals {
		if original.existed {
			if err := writeSiteFile(original.path, original.data, original.mode); err != nil {
				return err
			}
		} else if err := os.Remove(original.path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func nginxCommand(ctx context.Context, args ...string) error {
	output, err := exec.CommandContext(ctx, "nginx", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx %s: %w: %s", strings.Join(args, " "), err, output)
	}
	return nil
}

func (s *SettingService) ApplyNginxSite(ctx context.Context, v NginxSite) error {
	nginxSiteMu.Lock()
	defer nginxSiteMu.Unlock()
	// Do not take a backup while site files are being replaced.
	fullBackupMu.Lock()
	defer fullBackupMu.Unlock()
	configPath, index, err := validateSitePaths(v)
	if err != nil {
		return err
	}
	if len(v.Config) == 0 || len(v.Config) > 1<<20 || len(v.HTML) > 4<<20 {
		return fmt.Errorf("configuration is required (max 1 MiB); HTML max 4 MiB")
	}
	if err = nginxCommand(ctx, "-t"); err != nil {
		return err
	}
	originals := []siteOriginal{}
	for _, path := range []string{configPath, index} {
		original, err := rememberSiteFile(path)
		if err != nil {
			return err
		}
		originals = append(originals, original)
	}
	rollback := func(cause error) error {
		if restoreErr := restoreSiteFiles(originals); restoreErr != nil {
			return fmt.Errorf("%w; rollback failed: %v", cause, restoreErr)
		}
		return cause
	}
	if err = writeSiteFile(index, []byte(v.HTML), originals[1].mode); err != nil {
		return rollback(err)
	}
	if err = writeSiteFile(configPath, []byte(v.Config), originals[0].mode); err != nil {
		return rollback(err)
	}
	if err = nginxCommand(ctx, "-t"); err != nil {
		return rollback(err)
	}
	if err = nginxCommand(ctx, "-s", "reload"); err != nil {
		return rollback(err)
	}
	// Persist paths only; the website itself remains in the usual nginx directories.
	paths, err := json.Marshal(NginxSite{ConfigPath: v.ConfigPath, Root: v.Root})
	if err != nil {
		return err
	}
	if err = s.saveSetting("fork.nginxSite", string(paths)); err != nil {
		rollbackErr := rollback(err)
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if reloadErr := nginxCommand(rollbackCtx, "-s", "reload"); reloadErr != nil {
			return fmt.Errorf("%w; reload restored configuration: %v", rollbackErr, reloadErr)
		}
		return rollbackErr
	}
	return nil
}
