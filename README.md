# 3X-UI — martynov-dev fork

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

A fork of [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui), a panel for managing Xray/VPN. Its purpose is to remove advertising, back up the panel together with nginx and the VPN cover website, and configure that website through the panel.

Based on **3X-UI v3.10.2**, commit [608ef4e4010bf83e2e78328e08f77efc051ad4d7](https://github.com/MHSanaei/3x-ui/commit/608ef4e4010bf83e2e78328e08f77efc051ad4d7). This identifies the upstream version from which development of the fork started.

## Changes from upstream

- Removed sponsor advertisements, the sponsors page and advertising requests.
- Added manual and scheduled full backups: configurable intervals, local retention, additional source paths and optional SSH delivery to another device.
- Added nginx configuration and `index.html` editing, an HTTP template, `nginx -t` validation and file rollback if applying changes fails.
- Installation and panel updates use releases from **martynov-dev/3x-ui**. Upstream source changes can be merged separately.

## Installation and updates

The fork must first have a successful **Release 3X-UI** build in GitHub Actions and a published release containing binaries. The rolling build of `main` uses `dev-latest`:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/install.sh) dev-latest
```

To update an existing installation to this fork:

```bash
XUI_UPDATE_TAG=dev-latest bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/update.sh)
```

Back up before updating. Stable tags must exist in this fork; upstream releases are not substituted automatically. The panel’s Dev channel also checks this fork.

## Full backups

In the backups and website section, choose the interval and local retention, then enable scheduling. It is disabled by default, runs while the panel is running and is checked every minute. The first backup runs after enabling it; overdue work resumes after a restart.

Archives are stored in `/etc/x-ui/full-backups` or `$XUI_DB_FOLDER/full-backups` (directory `0700`, archives `0600`). They contain a consistent SQLite snapshot (`database/x-ui.db`) or PostgreSQL dump (`database/x-ui.dump`), panel files and service settings, configured nginx/site/certificate paths, symlink targets and `manifest.json` with the version, date, sources and missing paths. Static nginx paths are also discovered through `nginx -T`.

This is an application backup, not an OS image. Logs, old archives and live SQLite WAL/SHM files are excluded. Add paths using nginx variables and files from other services manually. Read errors on existing files abort the backup.

For SSH delivery, configure the host, port, user, existing destination directory and private key path **on the panel server**. The panel server needs `scp`, `ssh`, key authentication and a verified destination in `known_hosts`. The destination needs an scp-compatible server and POSIX `mv` (Linux/WSL, a router or another server); a native Windows SSH shell is unsupported. An empty host means local backups only. Configure remote retention on the destination. Incomplete uploads end in `.partial`; upload failures are reported and the local archive is kept.

Archives contain passwords and private keys. SSH encrypts transport; the archive on disk is not encrypted.

### Restoring

Full restoration is manual. Do not use the database import button for a full archive. Inspect `manifest.json`, install compatible panel/nginx versions and stop the panel. Restore `files/` to the corresponding absolute paths, preserving ownership, permissions and symlinks. Restore SQLite with the panel stopped, or restore the PostgreSQL dump. Run `nginx -t` before starting services. Test restoration on a separate installation first.

## nginx website

nginx must already be installed and running. Select a discovered configuration or a path under `/etc/nginx` and a website directory under `/var/www` or `/usr/share/nginx/html`. Load existing files before editing. The main `nginx.conf` is not edited.

The HTTP template listens on port 80; point DNS to the server and allow access. HTTPS requires existing certificates and suitable directives. Configure a matching local port for Xray fallback yourself. Certificate issuance and Xray inbound configuration are not automated.

The current and proposed configurations are validated. A write, validation or reload failure restores the previous files. A `sites-available` configuration needs an active `sites-enabled` link or `include`. Avoid domain/port conflicts. Only `index.html` is edited; other website assets remain on disk and are backed up with the website directory.

## Merging upstream changes

Commit local changes and start with a clean working tree. Add the upstream remote once, then fetch and merge:

```bash
git remote add upstream https://github.com/MHSanaei/3x-ui.git
git fetch upstream
git switch main
git merge upstream/main
```

Resolve conflicts, run project checks and push to your fork. Merging preserves commit history, but compatibility with future upstream changes needs testing. Do not reset the fork branch to upstream. Updating a running installation from a fork release is a separate step.

## Source and license

Original project: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). License: [GPL-3.0](LICENSE). Original authorship is retained. The Go module keeps its upstream name for compatibility. General upstream documentation: [docs.sanaei.dev](https://docs.sanaei.dev).
