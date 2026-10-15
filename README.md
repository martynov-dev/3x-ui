# 3X-UI — martynov-dev fork

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

A fork of [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui), a panel for managing Xray/VPN. Its purpose is to remove advertising, back up the panel together with nginx and the VPN cover website, and configure that website through the panel.

Based on **3X-UI v3.10.2**, commit [608ef4e4010bf83e2e78328e08f77efc051ad4d7](https://github.com/MHSanaei/3x-ui/commit/608ef4e4010bf83e2e78328e08f77efc051ad4d7). This identifies the upstream version from which development of the fork started.

> [!IMPORTANT]
> This project is intended for personal use only. Please do not use it for illegal purposes or in a production environment.

## Changes from upstream

- Removed sponsor advertisements, the sponsors page and advertising requests.
- Added manual and scheduled full backups: configurable intervals, local retention, additional source paths and optional SSH delivery to another device.
- Added nginx configuration and `index.html` editing, an HTTP template, `nginx -t` validation and file rollback if applying changes fails.
- Installation and panel updates use releases from **martynov-dev/3x-ui**. Upstream source changes can be merged separately.

## Installation and updates

Before installation, publish a release containing the required binaries and checksums. GitHub Actions **Release 3X-UI** can build and publish them, or you can prepare and upload them manually. The rolling build of `main` uses `dev-latest`:

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

## Base panel reference

The following sections retain the useful reference material from the base version. Fork-specific changes and installation commands are described above. Container images must be built locally or published from this fork; do not use an upstream image to test fork changes.

### Features

- **Multi-protocol inbounds** — VLESS, VMess, Trojan, Shadowsocks, WireGuard, AmneziaWG, TUIC v5, Hysteria2, MTProto, HTTP, SOCKS (Mixed), Dokodemo-door / Tunnel, and TUN.
- **Modern transports & security** — TCP (Raw), mKCP, WebSocket, gRPC, HTTPUpgrade, and XHTTP, secured with TLS, XTLS, and REALITY.
- **AmneziaWG built in** — DPI-resistant WireGuard runs inside the panel on a userspace network stack, with no kernel module, DKMS, or extra packages to install.
- **Native TUIC v5 server** — In-process Go QUIC server with Xray routing and per-client traffic accounting; BBR and New Reno are available server-side. CUBIC is preserved in the client profile but currently falls back to New Reno on the server.
- **MTProto proxies** — per-client FakeTLS secrets, ad-tags, and quotas, applied live without dropping existing connections.
- **Fallbacks** — serve multiple protocols on a single port (e.g. VLESS and Trojan on 443) using Xray's fallback support.
- **Per-client management** — traffic quotas, expiry dates, IP limits with trusted-address exemptions, HWID device limits, scheduled renewal cycles, live online status, and one-click share links, QR codes, and subscriptions.
- **Traffic statistics** — per inbound, per client, and per outbound, with reset controls.
- **Multi-node support** — manage and scale across multiple servers from a single panel, including cloning inbounds onto other nodes.
- **Outbound & routing** — WARP, NordVPN, PIA, custom routing rules, load balancers with balancer-to-balancer fallback, and outbound proxy chaining. Bundled geosite and geoip categories are browsable straight from the rule editor.
- **Built-in subscription server** — raw, JSON, and Clash output, auto-selected from the client's User-Agent, plus [custom page templates](docs/custom-subscription-templates.md).
- **Telegram and Discord bots** for remote monitoring and management.
- **RESTful API** with scoped, optionally expiring tokens and an in-panel API reference.
- **Installable panel (PWA)** — pin 3X-UI to a desktop or phone home screen.
- **Flexible storage** — SQLite (default) or PostgreSQL.
- **13 UI languages** with dark and light themes.
- **Fail2ban integration** for enforcing per-client IP limits.

### Screenshots

<details>
<summary>Click to expand</summary>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/01-overview-dark.png">
  <img alt="Overview" src="./media/01-overview-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/02-add-inbound-dark.png">
  <img alt="Inbounds" src="./media/02-add-inbound-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/03-add-client-dark.png">
  <img alt="Add client" src="./media/03-add-client-light.png">
</picture>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./media/05-add-nodes-dark.png">
  <img alt="Configs" src="./media/05-add-nodes-light.png">
</picture>

</details>

### Quick Start

During installation a random username, password, and access path are generated. After installation, run `x-ui` to open the management menu, where you can start/stop the service, view or reset your login credentials, manage SSL certificates, and more.

Every release asset is published with a `.sha256` sum next to it. Both `install.sh` and the updater verify the archive against that sum and abort on a mismatch.

For full documentation — installation, configuration, operations, and the complete API reference — visit **[docs.sanaei.dev](https://docs.sanaei.dev)**.

#### Unattended install

The installer also runs **non-interactively** for cloud-init.
Set `XUI_NONINTERACTIVE=1` (or pipe with no TTY) and it installs end-to-end with
zero prompts, generating random credentials and writing them to
`/etc/x-ui/install-result.env`. See [`deploy/`](deploy/) for:

- [Cloud-init user-data](deploy/cloud-init/) — unattended install on any cloud (Hetzner/AWS/DO/Vultr/GCP/Azure/Oracle)
- [Hetzner Cloud notes](deploy/marketplace/hetzner/) — cloud-init deployment on Hetzner

### Supported Platforms

**Operating systems:** Ubuntu, Debian, Armbian, Fedora, CentOS, RHEL, AlmaLinux, Rocky Linux, Oracle Linux, Amazon Linux, Virtuozzo, Arch, Manjaro, Parch, openSUSE (Tumbleweed / Leap), Alpine, and Windows.

**Architectures:** `amd64` · `386` · `arm64` (aarch64) · `armv7` · `armv6` · `armv5` · `s390x`.

### Database Options

3X-UI supports two backends, chosen during the install:

- **SQLite** (default) — a single file at `/etc/x-ui/x-ui.db`. Zero setup, ideal for small and medium deployments.
- **PostgreSQL** — recommended for high client counts or multi-node setups. The installer can install PostgreSQL locally for you, or accept a DSN to an existing server.

At runtime the backend is selected via environment variables (the installer writes these to `/etc/default/x-ui` for you):

```
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable
```

#### Migrating an existing SQLite install to PostgreSQL

```bash
x-ui migrate-db --dsn "postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable"
# then set XUI_DB_TYPE and XUI_DB_DSN in /etc/default/x-ui and restart:
systemctl restart x-ui
```

The source SQLite file is left untouched; remove it manually once you have verified the new backend.

#### Docker

The default `docker compose up -d` keeps using SQLite. To run with the bundled PostgreSQL service, uncomment the two `XUI_DB_*` env lines in `docker-compose.yml` and start with the profile:

```bash
docker compose --profile postgres up -d
```

The image bundles Fail2ban (enabled by default) to enforce per-client **IP limits**. Fail2ban bans offenders with `iptables`, which requires the `NET_ADMIN` capability. `docker-compose.yml` already grants it via `cap_add`; if you start the container with `docker run` instead, add the capabilities yourself, otherwise bans are logged but never applied:

```bash
docker run -d --cap-add=NET_ADMIN --cap-add=NET_RAW ... ghcr.io/martynov-dev/3x-ui
```

### Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| `XUI_DB_TYPE` | Database backend: `sqlite` or `postgres` | `sqlite` |
| `XUI_DB_DSN` | PostgreSQL connection string (when `XUI_DB_TYPE=postgres`) | — |
| `XUI_DB_FOLDER` | Directory for the SQLite database file | `/etc/x-ui` |
| `XUI_DB_MAX_OPEN_CONNS` | Maximum open connections (PostgreSQL pool) | — |
| `XUI_DB_MAX_IDLE_CONNS` | Maximum idle connections (PostgreSQL pool) | — |
| `XUI_INIT_WEB_BASE_PATH` | The initial URI path for the web panel | `/` |
| `XUI_ENABLE_FAIL2BAN` | Enable Fail2ban-based IP-limit enforcement | `true` |
| `XUI_LOG_LEVEL` | Log verbosity (`debug`, `info`, `warning`, `error`) | `info` |
| `XUI_DEBUG` | Enable debug mode | `false` |
| `XUI_TUNNEL_HEALTH_MONITOR` | Enable the tunnel health monitor (probes a URL and restarts xray after repeated failures; a restart drops all clients) | `false` |
| `XUI_TUNNEL_HEALTH_PROXY` | Proxy the probe is sent through; point it at a local xray inbound so the probe tests the tunnel (e.g. `socks5://127.0.0.1:1080`). Empty means the probe only checks host connectivity | — |
| `XUI_TUNNEL_HEALTH_URL` | URL probed for tunnel health | `https://www.cloudflare.com/cdn-cgi/trace` |
| `XUI_TUNNEL_HEALTH_INTERVAL` | Interval between probes | `30s` |
| `XUI_TUNNEL_HEALTH_TIMEOUT` | Per-probe timeout | `10s` |
| `XUI_TUNNEL_HEALTH_FAILURES` | Consecutive failures before a restart is triggered | `3` |
| `XUI_TUNNEL_HEALTH_COOLDOWN` | Minimum delay between consecutive restarts | `5m` |
| `NODE_TOKEN_ENCRYPTION` | Encryption at rest for node API tokens: `off`, `migration`, or `required` (note: no `XUI_` prefix) | `off` |
| `XUI_NODE_TOKEN_KEY_FILE` | JSON keyring (mode `0600`) holding the active key id and its base64 32-byte keys | `/etc/x-ui/node_token_key.json` |
| `XUI_NODE_TOKEN_KEY` | A single base64 32-byte key, used only when the key file cannot be loaded | — |

The complete list is on the [environment variables reference](https://docs.sanaei.dev/docs/reference/env-vars).

### Supported Languages

The panel UI is available in 13 languages:

English · فارسی · العربية · 中文（简体） · 中文（繁體） · Español · Русский · Українська · Türkçe · Tiếng Việt · 日本語 · Bahasa Indonesia · Português (Brasil)

### Contributing

Contributions are welcome. Please read the [Contributing Guide](/CONTRIBUTING.md) before opening an issue or pull request.

### Community Tools

Tools and integrations built by the community around 3x-ui.

- [terraform-provider-3x-ui](https://github.com/batonogov/terraform-provider-threexui) (License: **MIT**): _Manage inbounds, clients, panel settings, and Xray configuration as code with Terraform / OpenTofu._
- [3X-UI Manager](https://github.com/yukh975/3X-UI-Manager) (License: **MIT**): _Native Android client for 3x-ui — dashboard, inbounds, clients with QR sharing, nodes and multi-panel management. Available on F-Droid._

## Source and license

Original project: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). License: [GPL-3.0](LICENSE). Original authorship is retained. The Go module keeps its upstream name for compatibility. General upstream documentation: [docs.sanaei.dev](https://docs.sanaei.dev).
