# Full backups: computer, router or server

[English](full-backups.md) | [Русский — подробная инструкция](full-backups.ru.md)

## Download without a terminal

Open **Backups and website → Full backups**, click **Back up now**, wait, then click the archive name under **Saved archives**. The browser downloads a `.tar.gz` to your computer. It contains passwords and private keys; keep it private. This is a manual download, not automatic delivery to a computer.

## Schedule

Enable **Automatic backups**, choose an interval in hours (`24` = daily, `168` = weekly), set local retention and click **Save**. The scheduler checks once a minute while the panel runs. With no previous attempt, the first backup starts within a minute; otherwise the interval is measured from the previous attempt. Click **Refresh** to see new archives.

The database and panel files are always included. Default nginx, website and certificate paths are prefilled under **Additional paths**. Add custom directories if needed. This is an application backup, not an OS image.

## Automatic remote delivery

One-time SSH setup currently needs a terminal or an administrator. The panel form does not create keys or configure the destination.

1. On the **panel server**, install `openssh-client` and generate a dedicated passwordless key, for example `/root/.ssh/backup_ed25519`, using `ssh-keygen -t ed25519 -N '' -f /root/.ssh/backup_ed25519`. Create `/root/.ssh` with mode `700` first; do not overwrite an existing key.
2. On the **receiving Linux server**, install `openssh-server`, create a `backup` user and a writable directory such as `/srv/backups/3x-ui`. Add the panel key's `.pub` line to `/home/backup/.ssh/authorized_keys`, owned by `backup`; use directory mode `700` and file mode `600`.
3. On the panel server, connect once with `ssh -i /root/.ssh/backup_ed25519 backup@backup.example.com`. Compare the host fingerprint with a trusted reading on the destination before accepting it. This records the destination in `known_hosts`. Check that key authentication works without a prompt.
4. In the panel expand **Remote copy over SSH**. Enter the receiving address, user, SSH port, receiving directory and the private key path **on the panel server**. Save, then click **Back up now** and verify a finished `.tar.gz` on the destination.

The destination needs SFTP and a POSIX shell with `mv`. A `.partial` file becomes the final archive after delivery. A failed upload retains the local archive and reports an error. Local retention does not delete remote copies. Clear the host and save to disable delivery.

The panel server must be able to reach the destination. A home `192.168.x.x` address requires an existing shared network or VPN connection; it is not directly reachable from a public VPS. A receiving PC must be running during delivery. Native Windows OpenSSH is currently unsupported; use browser downloads or a prepared Linux VM.

For a **Banana Pi R4 running OpenWrt**, mount the SD card persistently in LuCI, install `openssh-sftp-server`, authorize the panel public key under SSH keys, and create a private backup directory on the mounted card. Verify that the directory is on the SD card before using it. Standard OpenWrt usually uses the `root` SSH user, so authorizing a key grants administrative access. Use an existing private VPN connection to reach it. See the [detailed Russian steps](full-backups.ru.md), [OpenWrt storage documentation](https://openwrt.org/docs/guide-user/storage/fstab) and [SFTP documentation](https://openwrt.org/docs/guide-user/services/nas/sftp.server).

For **Docker panels**, persist the private key and `known_hosts` inside the container, normally using a private `/root/.ssh` volume. The fork image includes the SSH client. Host nginx and website directories must be explicitly made available to the container; host services are not automatically managed.

## Verify restoration

Use a separate matching Ubuntu/Debian machine to run `setup.sh --backup /path/full-backup.tar.gz --check`, then test actual restoration on a clean machine. Automatic restoration supports this fork's full Linux SQLite backups. PostgreSQL, Docker/Alpine, external firewall rules and DNS need separate handling. See the [README](../README.md) for restoration commands and publication prerequisites.

### Verified VM scenario

On October 20, 2026, installation and restoration were tested between two separate clean Ubuntu 24.04.5 LTS VMs (amd64, SQLite) running actual systemd:

- Installation through `setup.sh`, panel/nginx startup and panel login.
- Website creation through the panel API under the shipped systemd service restrictions, HTTP and a certificate stored outside the default paths.
- Full backup and SSH delivery with matching SHA256.
- `--check` without installing the panel/nginx, followed by full restoration on the second VM.
- Preserved login, schedule/destination settings, SSH identity, executable, certificate and test Xray inbound; working HTTP/HTTPS and an active Xray listening port.
- Another successful SSH backup after restoration without reconfiguring SSH.
- An update through `update.sh` retaining login, settings, website and working SSH delivery.
- Automatic backup creation and delivery by the restored scheduler: only the test database's last-attempt timestamp was made overdue; the panel's actual scheduler executed the backup.

During an update, the existing `update.sh` replaces a very short panel path (such as `/`) with a random path and prints the new URL. Account migration does not guarantee that such a URL remains unchanged.

The unpublished build used a local release archive; only fork file downloads were redirected to a fixture HTTP server. Fetching a published GitHub release was not covered. Debian, other OS versions/architectures, PostgreSQL and live VPN traffic restoration were also outside this run. Logs and the test archive remain locally in `.cache/vm-test/`; the archive contains a test private key.
