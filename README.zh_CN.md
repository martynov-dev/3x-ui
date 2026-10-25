# 3X-UI — martynov-dev 分支

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

本项目是 Xray/VPN 管理面板 [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) 的分支。目标是移除广告，备份面板、nginx 和 VPN 伪装网站，并通过面板配置网站。

基于 **3X-UI v3.10.2**，起始提交为 [608ef4e4010bf83e2e78328e08f77efc051ad4d7](https://github.com/MHSanaei/3x-ui/commit/608ef4e4010bf83e2e78328e08f77efc051ad4d7)。这是本分支开始开发时采用的上游版本。

> [!IMPORTANT]
> 本项目仅供个人使用。请勿将其用于非法目的，也请勿在生产环境中使用。

## 与原项目的区别

- 移除了赞助商广告、赞助商页面和加载广告的请求。
- 新增手动和定时完整备份，可配置间隔、本地保留数量、额外路径以及通过 SSH 发送到其他设备。
- 新增 nginx 配置和 `index.html` 编辑、HTTP 模板、`nginx -t` 验证，以及应用失败时的文件回滚。
- 安装和面板更新使用 **martynov-dev/3x-ui** 的发布版本。上游源码更新需要单独合并。

## 安装和更新

安装前，请发布包含所需二进制文件和校验和的版本。可使用 GitHub Actions **Release 3X-UI** 构建和发布，也可手动准备并上传。`main` 分支的滚动构建使用 `dev-latest`：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/setup.sh) dev-latest
```

将现有安装更新到本分支：

```bash
XUI_UPDATE_TAG=dev-latest bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/update.sh)
```

更新前请先备份。稳定版本标签必须已在本分支发布，不会自动替换为上游发布版本。面板的 Dev 通道也检查本分支。

## 完整备份

在备份与网站页面选择间隔和本地保留数量，然后启用定时备份。默认关闭，仅在面板运行时执行，每分钟检查一次。启用后执行首次备份，重启后执行已到期的任务。

存储位置为 `/etc/x-ui/full-backups` 或 `$XUI_DB_FOLDER/full-backups`（目录权限 `0700`，归档权限 `0600`）。内容包括一致的 SQLite 快照（`database/x-ui.db`）或 PostgreSQL 转储（`database/x-ui.dump`）、面板文件与服务设置、配置的 nginx/网站/证书路径、符号链接目标，以及记录版本、日期、来源和缺失路径的 `manifest.json`。还会通过 `nginx -T` 发现静态 nginx 路径。

这是应用备份，不是操作系统镜像。不包含日志、旧归档和活动 SQLite WAL/SHM 文件。使用 nginx 变量的路径和其他服务文件需要手动添加。读取现有文件失败会中止备份。

远程发送需要配置 SSH 地址、端口、用户、已存在的目标目录，以及**面板服务器上**的私钥路径。面板服务器需要 `scp`、`ssh`、密钥认证及 `known_hosts` 中已验证的目标主机密钥。目标需要兼容 scp 的服务器和 POSIX `mv` 命令（Linux/WSL、路由器或其他服务器）；不支持原生 Windows SSH shell。地址为空时仅本地保存。远程保留策略需要在目标设备上配置。未完成的上传以 `.partial` 结尾；上传失败会显示错误并保留本地归档。

归档包含密码和私钥。SSH 加密传输过程，但磁盘上的归档本身未加密。

### 恢复

Ubuntu/Debian 与 SQLite 可使用同一个脚本安装或恢复，请以 root 运行。完整 Linux 备份必须来自相同 CPU 架构和发行版版本。脚本验证归档和 SQLite 快照，安装依赖，恢复归档中的面板程序、数据库、nginx、网站与证书，并启动服务。面板地址和登录凭据保持为备份中的值。--check 仅检查不应用；--replace 明确允许替换现有安装。应用失败时回滚文件；回滚失败时原文件保留在 /var/backups/x-ui-restore-*。PostgreSQL 和 Docker 备份目前需要手动恢复。外部防火墙、DNS 和外部服务不会恢复。只恢复可信归档。使用下载地址前，脚本必须已在本分支发布。

```bash
curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/setup.sh -o setup.sh

sudo bash setup.sh --backup /root/full-backup.tar.gz --check

sudo bash setup.sh --backup /root/full-backup.tar.gz
```

## nginx 网站

nginx 必须已安装并运行。选择检测到的配置，或指定 `/etc/nginx` 内的路径和 `/var/www` 或 `/usr/share/nginx/html` 内的网站目录。编辑前先加载现有文件。编辑器不会修改主 `nginx.conf`。

HTTP 模板监听 80 端口；请将 DNS 指向服务器并允许访问。HTTPS 需要现有证书和相应指令。Xray fallback 需要手动配置匹配的本地端口。不会自动签发证书或配置 Xray 入站。

会验证当前配置和新配置。写入、验证或 reload 失败时恢复原文件。`sites-available` 中的配置必须通过有效的 `sites-enabled` 链接或 `include` 启用。避免域名和端口冲突。编辑器只修改 `index.html`；其他资源保留在磁盘上，并随网站目录一起备份。

## 合并上游更新

先提交本地修改并确保工作区干净。upstream 远程仓库只需添加一次，然后获取并合并更新：

```bash
git remote add upstream https://github.com/MHSanaei/3x-ui.git
git fetch upstream
git switch main
git merge upstream/main
```

解决冲突，运行项目检查，然后推送到自己的分支。合并会保留提交历史，但与未来上游变更的兼容性仍需测试。不要将本分支重置到 upstream。通过本分支发布版本更新正在运行的面板是另一个独立步骤。

## 基础面板参考

以下保留基础版本中的实用参考内容。本分支的改动和安装命令见上文。容器镜像需要在本地构建或从本分支发布；上游镜像不包含本分支的改动。

### 功能特性

- **多协议入站** — VLESS、VMess、Trojan、Shadowsocks、WireGuard、AmneziaWG、TUIC v5、Hysteria2、MTProto、HTTP、SOCKS (Mixed)、Dokodemo-door / Tunnel 和 TUN。
- **现代传输与安全** — TCP (Raw)、mKCP、WebSocket、gRPC、HTTPUpgrade 和 XHTTP，并通过 TLS、XTLS 和 REALITY 加密。
- **内置 AmneziaWG** — 抗 DPI 的 WireGuard 直接在面板内的用户态网络栈上运行，无需内核模块、DKMS 或额外软件包。
- **内置 TUIC v5** — 基于 QUIC 的高性能代理，支持原生 UDP 中继流量统计、0-RTT 握手和 BBR 拥塞控制。
- **MTProto 代理** — 按客户端配置 FakeTLS 密钥、广告标签和配额，实时生效且不会断开已有连接。
- **回落 (Fallback)** — 通过 Xray 的 fallback 功能在单个端口上提供多种协议（例如在 443 端口上同时使用 VLESS 和 Trojan）。
- **按客户端管理** — 流量配额、到期日期、可豁免受信任地址的 IP 限制、HWID 设备数限制、定时续期周期、实时在线状态，以及一键分享链接、二维码和订阅。
- **流量统计** — 按入站、按客户端、按出站统计，并支持重置控制。
- **多节点支持** — 从单一面板管理并扩展到多台服务器，并可将入站克隆到其他节点。
- **出站与路由** — WARP、NordVPN、PIA、自定义路由规则、支持均衡器间回退的负载均衡器，以及出站代理链。内置的 geosite 与 geoip 分类可直接在规则编辑器中浏览。
- **内置订阅服务器** — 提供 raw、JSON 和 Clash 输出，可依据客户端 User-Agent 自动选择，并支持[自定义页面模板](docs/custom-subscription-templates.md)。
- **Telegram 和 Discord 机器人**，用于远程监控和管理。
- **RESTful API**，支持带作用域、可设置有效期的令牌，并提供面板内置的 API 参考文档。
- **可安装面板 (PWA)** — 将 3X-UI 固定到桌面或手机主屏幕。
- **灵活的存储** — SQLite（默认）或 PostgreSQL。
- **14 种界面语言**，支持深色和浅色主题。
- **Fail2ban 集成**，用于强制执行按客户端的 IP 限制。

### 截图

<details>
<summary>点击展开</summary>

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

### 快速开始

安装过程中会生成随机的用户名、密码和访问路径。安装完成后，运行 `x-ui` 打开管理菜单，您可以在其中启动/停止服务、查看或重置登录凭据、管理 SSL 证书等。

每个发布资源都会在其旁边附带一个 `.sha256` 校验和。`install.sh` 和更新程序都会据此校验压缩包，不匹配时中止。

完整文档（安装、配置、运维以及完整的 API 参考）请访问 **[docs.sanaei.dev](https://docs.sanaei.dev/zh)**。

#### 无人值守安装

安装程序也可以**非交互式**运行，适用于 cloud-init。
设置 `XUI_NONINTERACTIVE=1`（或在无 TTY 的情况下通过管道传入），它就会全程
零提示地完成端到端安装，生成随机凭据并写入
`/etc/x-ui/install-result.env`。请参阅 [`deploy/`](deploy/)：

- [Cloud-init user-data](deploy/cloud-init/) — 在任意云平台上无人值守安装（Hetzner/AWS/DO/Vultr/GCP/Azure/Oracle）
- [Hetzner Cloud 说明](deploy/marketplace/hetzner/) — 在 Hetzner 上基于 cloud-init 的部署

### 支持的平台

**操作系统：** Ubuntu、Debian、Armbian、Fedora、CentOS、RHEL、AlmaLinux、Rocky Linux、Oracle Linux、Amazon Linux、Virtuozzo、Arch、Manjaro、Parch、openSUSE (Tumbleweed / Leap)、Alpine 和 Windows。

**架构：** `amd64` · `386` · `arm64` (aarch64) · `armv7` · `armv6` · `armv5` · `s390x`。

### 数据库选项

3X-UI 支持两种后端，可在安装时选择：

- **SQLite**（默认）— 位于 `/etc/x-ui/x-ui.db` 的单个文件。无需配置，适合中小型部署。
- **PostgreSQL** — 推荐用于大量客户端或多节点设置。安装程序可以为您在本地安装 PostgreSQL，或接受指向现有服务器的 DSN。

运行时通过环境变量选择后端（安装程序会为您写入 `/etc/default/x-ui`）：

```
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable
```

#### 将现有的 SQLite 安装迁移到 PostgreSQL

```bash
x-ui migrate-db --dsn "postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable"
# 然后在 /etc/default/x-ui 中设置 XUI_DB_TYPE 和 XUI_DB_DSN 并重启：
systemctl restart x-ui
```

源 SQLite 文件保持不变；在确认新后端正常工作后，请手动删除它。

#### Docker

默认的 `docker compose up -d` 仍使用 SQLite。若要使用捆绑的 PostgreSQL 服务运行，请取消注释 `docker-compose.yml` 中的两行 `XUI_DB_*` 环境变量，并使用该 profile 启动：

```bash
docker compose --profile postgres up -d
```

该镜像捆绑了 Fail2ban（默认启用），用于强制执行按客户端的 **IP 限制**。Fail2ban 使用 `iptables` 封禁违规者，这需要 `NET_ADMIN` 权限。`docker-compose.yml` 已通过 `cap_add` 授予该权限；如果您改用 `docker run` 启动容器，请自行添加这些权限，否则封禁只会被记录而永远不会生效：

```bash
docker run -d --cap-add=NET_ADMIN --cap-add=NET_RAW ... ghcr.io/martynov-dev/3x-ui
```

### 环境变量

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `XUI_DB_TYPE` | 数据库后端：`sqlite` 或 `postgres` | `sqlite` |
| `XUI_DB_DSN` | PostgreSQL 连接字符串（当 `XUI_DB_TYPE=postgres` 时） | — |
| `XUI_DB_FOLDER` | SQLite 数据库文件所在目录 | `/etc/x-ui` |
| `XUI_DB_MAX_OPEN_CONNS` | 最大打开连接数（PostgreSQL 连接池） | — |
| `XUI_DB_MAX_IDLE_CONNS` | 最大空闲连接数（PostgreSQL 连接池） | — |
| `XUI_INIT_WEB_BASE_PATH` | Web 面板的初始 URI 路径 | `/` |
| `XUI_ENABLE_FAIL2BAN` | 启用基于 Fail2ban 的 IP 限制 | `true` |
| `XUI_LOG_LEVEL` | 日志级别（`debug`、`info`、`warning`、`error`） | `info` |
| `XUI_DEBUG` | 启用调试模式 | `false` |
| `XUI_TUNNEL_HEALTH_MONITOR` | 启用隧道健康监控（探测某个 URL，在连续多次失败后重启 xray；重启会断开所有客户端） | `false` |
| `XUI_TUNNEL_HEALTH_PROXY` | 探测请求所经过的代理；将其指向本地 xray 入站，使探测能够测试隧道（例如 `socks5://127.0.0.1:1080`）。留空表示探测仅检查主机连通性 | — |
| `XUI_TUNNEL_HEALTH_URL` | 用于检测隧道健康状况的探测 URL | `https://www.cloudflare.com/cdn-cgi/trace` |
| `XUI_TUNNEL_HEALTH_INTERVAL` | 两次探测之间的间隔 | `30s` |
| `XUI_TUNNEL_HEALTH_TIMEOUT` | 单次探测的超时时间 | `10s` |
| `XUI_TUNNEL_HEALTH_FAILURES` | 触发重启前的连续失败次数 | `3` |
| `XUI_TUNNEL_HEALTH_COOLDOWN` | 两次连续重启之间的最小间隔 | `5m` |
| `NODE_TOKEN_ENCRYPTION` | 节点 API 令牌的静态加密：`off`、`migration` 或 `required`（注意：无 `XUI_` 前缀） | `off` |
| `XUI_NODE_TOKEN_KEY_FILE` | JSON 密钥环（权限 `0600`），包含活动密钥 ID 及其 base64 编码的 32 字节密钥 | `/etc/x-ui/node_token_key.json` |
| `XUI_NODE_TOKEN_KEY` | 单个 base64 编码的 32 字节密钥，仅在无法加载密钥文件时使用 | — |

完整列表请参阅[环境变量参考](https://docs.sanaei.dev/zh/docs/reference/env-vars)。

### 支持的语言

面板界面提供 14 种语言：

English · فارسی · العربية · 中文（简体） · 中文（繁體） · Español · Русский · Українська · Türkçe · Tiếng Việt · 日本語 · Bahasa Indonesia · Português (Brasil) · Slovenčina

### 贡献

欢迎贡献。在提交 issue 或 pull request 之前，请阅读[贡献指南](/CONTRIBUTING.md)。

### 社区工具

社区围绕 3x-ui 构建的工具和集成。

- [terraform-provider-3x-ui](https://github.com/batonogov/terraform-provider-threexui) (许可证: **MIT**): _使用 Terraform / OpenTofu 通过代码管理入站、客户端、面板设置和 Xray 配置。_
- [3X-UI Manager](https://github.com/yukh975/3X-UI-Manager) (许可证: **MIT**): _3x-ui 的原生 Android 客户端 — 仪表板、入站、带二维码分享的客户端、节点以及多面板管理。可在 F-Droid 获取。_

## 来源与许可证

原项目：[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui)。许可证：[GPL-3.0](LICENSE)。保留原作者信息。为保持代码兼容，Go 模块名称保持不变。上游通用文档：[docs.sanaei.dev](https://docs.sanaei.dev)。
