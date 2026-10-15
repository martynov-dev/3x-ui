# 3X-UI — martynov-dev 分支

[English](README.md) | [فارسی](README.fa_IR.md) | [العربية](README.ar_EG.md) | [中文](README.zh_CN.md) | [Español](README.es_ES.md) | [Русский](README.ru_RU.md) | [Türkçe](README.tr_TR.md)

本项目是 Xray/VPN 管理面板 [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) 的分支。目标是移除广告，备份面板、nginx 和 VPN 伪装网站，并通过面板配置网站。

基于 **3X-UI v3.10.2**，起始提交为 [608ef4e4010bf83e2e78328e08f77efc051ad4d7](https://github.com/MHSanaei/3x-ui/commit/608ef4e4010bf83e2e78328e08f77efc051ad4d7)。这是本分支开始开发时采用的上游版本。

## 与原项目的区别

- 移除了赞助商广告、赞助商页面和加载广告的请求。
- 新增手动和定时完整备份，可配置间隔、本地保留数量、额外路径以及通过 SSH 发送到其他设备。
- 新增 nginx 配置和 `index.html` 编辑、HTTP 模板、`nginx -t` 验证，以及应用失败时的文件回滚。
- 安装和面板更新使用 **martynov-dev/3x-ui** 的发布版本。上游源码更新需要单独合并。

## 安装和更新

安装前，本分支的 GitHub Actions 必须成功完成 **Release 3X-UI** 构建，并发布包含二进制文件的版本。`main` 分支的滚动构建使用 `dev-latest`：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/martynov-dev/3x-ui/main/install.sh) dev-latest
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

完整恢复需要手动进行。不要使用数据库导入按钮导入完整归档。检查 `manifest.json`，安装兼容的面板/nginx 版本并停止面板。将 `files/` 恢复到对应绝对路径，保留所有者、权限和符号链接。在面板停止时恢复 SQLite 快照，或恢复 PostgreSQL 转储。启动服务前运行 `nginx -t`。请先在独立安装上测试恢复。

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

## 来源与许可证

原项目：[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui)。许可证：[GPL-3.0](LICENSE)。保留原作者信息。为保持代码兼容，Go 模块名称保持不变。上游通用文档：[docs.sanaei.dev](https://docs.sanaei.dev)。
