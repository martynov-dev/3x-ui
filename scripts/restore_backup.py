#!/usr/bin/env python3
"""Restore trusted full SQLite backups on matching Ubuntu/Debian systemd hosts."""

import argparse
from contextlib import closing
import json
import os
import platform
import re
import shlex
import shutil
import sqlite3
import subprocess
import sys
import tarfile
import tempfile
import time
from pathlib import Path, PurePosixPath


def absolute_path(value):
    if not isinstance(value, str) or not value.startswith("/"):
        raise ValueError("Expected an absolute Linux path")
    if value == "/" or ".." in value.split("/") or "\x00" in value:
        raise ValueError("Invalid restore path")
    return str(PurePosixPath(value))


def os_release(text):
    values = {}
    for line in text.splitlines():
        if line.startswith(("ID=", "VERSION_ID=")):
            key, value = line.split("=", 1)
            parts = shlex.split(value)
            values[key] = parts[0] if parts else ""
    return values


def inspect_archive(archive, staging):
    """Validate all members before writing; never extract through archive symlinks."""
    members = archive.getmembers()
    names = set()
    links = set()
    for member in members:
        name = member.name.rstrip("/")
        parts = name.split("/")
        if not name or name.startswith("/") or any(p in ("", ".", "..") for p in parts):
            raise ValueError("Unsafe archive member: " + member.name)
        if name in names:
            raise ValueError("Duplicate archive member: " + name)
        names.add(name)
        if name == "manifest.json":
            if not member.isfile() or member.size > 1024 * 1024:
                raise ValueError("Invalid manifest")
        elif name.startswith("files/"):
            if not (member.isfile() or member.isdir() or member.issym()):
                raise ValueError("Unsupported file type: " + name)
            if member.issym():
                if not member.linkname or "\x00" in member.linkname:
                    raise ValueError("Invalid symlink")
                links.add(name)
        elif name == "database/x-ui.db":
            if not member.isfile():
                raise ValueError("Invalid SQLite snapshot")
        elif name == "database/x-ui.dump":
            raise ValueError("Automatic restore currently supports SQLite backups only; PostgreSQL requires manual dump restoration")
        else:
            raise ValueError("Unexpected archive member: " + name)
    for name in names:
        if any(str(parent) in links for parent in PurePosixPath(name).parents):
            raise ValueError("Archive writes through a symlink: " + name)
    if not {"manifest.json", "database/x-ui.db"}.issubset(names):
        raise ValueError("Not a full SQLite backup from this fork")
    with archive.extractfile("manifest.json") as source:
        manifest = json.load(source)
    if not isinstance(manifest, dict) or not isinstance(manifest.get("version"), str):
        raise ValueError("Invalid manifest fields")
    if not isinstance(manifest.get("osRelease", ""), str):
        raise ValueError("Invalid operating system metadata")
    if manifest.get("formatVersion", 0) not in (0, 1):
        raise ValueError("Unsupported backup format version")
    if manifest.get("os", "linux") != "linux":
        raise ValueError("Only Linux backups can be restored by this script")
    if manifest.get("dbType", "sqlite") != "sqlite":
        raise ValueError("Only SQLite is supported for automatic restoration")
    packages = manifest.get("packages", [])
    if not isinstance(packages, list) or any(
        not isinstance(p, str) or not re.fullmatch(r"(?:nginx[a-z0-9.+-]*|libnginx-mod-[a-z0-9.+-]+)(?::[a-z0-9]+)?", p)
        for p in packages
    ):
        raise ValueError("Invalid nginx package list")
    executable = absolute_path(manifest.get("executable", "/usr/local/x-ui/x-ui"))
    db_path = absolute_path(manifest.get("dbPath", "/etc/x-ui/x-ui.db"))
    if "files" + executable not in names:
        raise ValueError("Panel executable missing from backup: " + executable)
    if "files" + db_path in names:
        raise ValueError("Archive contains a live database as well as a snapshot")
    if not any(n in names for n in ("files/etc/systemd/system/x-ui.service", "files/usr/lib/systemd/system/x-ui.service")):
        raise ValueError("Backup does not contain an x-ui systemd service")
    if any(n.startswith("files/etc/nginx/") for n in names) and "files/etc/nginx/nginx.conf" not in names:
        raise ValueError("Backup contains an incomplete nginx configuration")
    needed = sum(m.size for m in members if m.isfile())
    if shutil.disk_usage(staging).free < needed + 64 * 1024 * 1024:
        raise ValueError("Not enough free space to unpack the backup")
    # Write regular files only. Symlinks are created at the destination later.
    for member in members:
        if member.isfile():
            path = staging / member.name
            path.parent.mkdir(parents=True, exist_ok=True)
            with archive.extractfile(member) as source, path.open("wb") as target:
                shutil.copyfileobj(source, target)
            path.chmod(member.mode & 0o777)
    snapshot = staging / "database/x-ui.db"
    with closing(sqlite3.connect(snapshot.as_uri() + "?mode=ro", uri=True)) as db:
        if db.execute("PRAGMA quick_check").fetchone()[0] != "ok":
            raise ValueError("SQLite snapshot integrity check failed")
    return manifest, members, executable, db_path


def check_platform(manifest, executable):
    if platform.system() != "Linux":
        raise ValueError("Run restoration on an Ubuntu/Debian Linux server")
    current = os_release(Path("/etc/os-release").read_text())
    if current.get("ID") not in ("ubuntu", "debian"):
        raise ValueError("Automatic restoration supports Ubuntu/Debian only")
    source = os_release(manifest.get("osRelease", ""))
    if source and source != current:
        raise ValueError("Use the same distribution and VERSION_ID as the backup server")
    # Check the actual ELF binary too, including archives without new metadata.
    with executable.open("rb") as binary:
        header = binary.read(20)
    expected = {"x86_64": 62, "aarch64": 183, "armv7l": 40, "armv6l": 40, "i686": 3, "s390x": 22}.get(platform.machine())
    if len(header) < 20 or header[:4] != b"\x7fELF" or header[5] not in (1, 2):
        raise ValueError("Backup executable is not a Linux ELF binary")
    machine = int.from_bytes(header[18:20], "little" if header[5] == 1 else "big")
    if expected is None or machine != expected:
        raise ValueError("Backup CPU architecture differs from this server")
    if not Path("/run/systemd/system").is_dir():
        raise ValueError("A running systemd host is required; do not run inside a plain Docker container")


def run(*args):
    subprocess.run(args, check=True)


def active(service):
    return subprocess.run(["systemctl", "is-active", "--quiet", service], check=False).returncode == 0


def apply_files(members, staging, root, rollback, changes):
    """Record overwritten files and new paths so failures can be rolled back."""
    def record(destination, old):
        changes.append((destination, old))
        journal = rollback / "paths.json.tmp"
        journal.write_text(json.dumps([[str(p), str(o) if o else None] for p, o in changes]))
        journal.replace(rollback / "paths.json")

    def prepare(destination):
        missing = []
        parent = destination.parent
        while not parent.exists():
            missing.append(parent)
            parent = parent.parent
        # An existing symlink parent could redirect a write outside its path.
        for parent in destination.parents:
            if parent == root:
                break
            if parent.is_symlink():
                raise ValueError("Destination parent is a symlink: " + str(parent))
        for parent in reversed(missing):
            parent.mkdir()
            record(parent, None)
        old = None
        if destination.exists() or destination.is_symlink():
            if destination.is_dir() and not destination.is_symlink():
                raise ValueError("Cannot replace a directory with a file: " + str(destination))
            old = rollback / str(len(changes))
            if destination.is_symlink():
                old.symlink_to(os.readlink(destination))
            else:
                shutil.copy2(destination, old)
                if hasattr(os, "chown"):
                    info = destination.stat()
                    os.chown(old, info.st_uid, info.st_gid)
        record(destination, old)
        if destination.is_symlink():
            destination.unlink()

    for member in sorted((m for m in members if m.isdir()), key=lambda m: len(m.name.split("/"))):
        destination = root / member.name[len("files/"):]
        # Reuse existing directories without changing their ownership or permissions.
        if destination.exists() and not destination.is_symlink():
            if not destination.is_dir():
                raise ValueError("Destination is not a directory: " + str(destination))
            continue
        prepare(destination)
        destination.mkdir(mode=member.mode & 0o777)
        if hasattr(os, "chown"):
            os.chown(destination, member.uid, member.gid)
    for member in members:
        if not member.name.startswith("files/") or member.isdir():
            continue
        destination = root / member.name[len("files/"):]
        prepare(destination)
        if member.issym():
            if destination.exists():
                destination.unlink()
            destination.symlink_to(member.linkname)
        else:
            shutil.copyfile(staging / member.name, destination)
            destination.chmod(member.mode & 0o777)
            os.utime(destination, (member.mtime, member.mtime))
        if hasattr(os, "chown"):
            os.chown(destination, member.uid, member.gid, follow_symlinks=False)
    return prepare


def undo_files(changes):
    for destination, old in reversed(changes):
        if destination.is_symlink() or destination.is_file():
            destination.unlink()
        elif destination.is_dir():
            destination.rmdir()
        if old is not None:
            if old.is_symlink():
                destination.symlink_to(os.readlink(old))
            else:
                shutil.copy2(old, destination)
                if hasattr(os, "chown"):
                    info = old.stat()
                    os.chown(destination, info.st_uid, info.st_gid)


def restore(args):
    with tempfile.TemporaryDirectory(prefix="x-ui-restore-") as temp:
        staging = Path(temp)
        with tarfile.open(args.backup, "r:gz") as archive:
            manifest, members, executable, db_path = inspect_archive(archive, staging)
        check_platform(manifest, staging / ("files" + executable))
        has_nginx = any(m.name.startswith("files/etc/nginx/") for m in members)
        print("Backup version:", manifest["version"])
        print("Panel executable:", executable)
        print("Database:", db_path)
        print("Files:", len(members), "nginx:", has_nginx)
        if manifest.get("missing"):
            print("Warning: backup records missing source paths; these cannot be restored.")
        if args.check:
            print("Validation passed; no services or destination files changed.")
            return
        if os.geteuid() != 0:
            raise ValueError("Run with sudo or as root")
        existing_paths = [executable, db_path, "/etc/systemd/system/x-ui.service", "/usr/lib/systemd/system/x-ui.service"]
        if has_nginx:
            existing_paths.append("/etc/nginx/nginx.conf")
        if not args.replace and any(Path(p).exists() for p in existing_paths):
            raise ValueError("Existing installation detected. Use --replace explicitly, or restore on a clean server")
        # Install dependencies, not a newer panel: the archive carries its exact executable.
        run("apt-get", "update")
        packages = ["ca-certificates", "curl", "openssl", "openssh-client", "fail2ban", "cron"]
        if has_nginx:
            packages.append("nginx")
            packages.extend(manifest.get("packages", []))
        env = dict(os.environ, DEBIAN_FRONTEND="noninteractive", NEEDRESTART_MODE="l")
        subprocess.run(["apt-get", "install", "-y", *packages], check=True, env=env)
        running = {service: active(service) for service in ("x-ui", "nginx")}
        changes = []
        # Keep originals outside the temporary extraction: failed rollback or a
        # killed process must not delete the only remaining copy of old files.
        rollback = Path(tempfile.mkdtemp(prefix="x-ui-restore-", dir="/var/backups"))
        print("Recovery files and path journal:", rollback, flush=True)
        try:
            for service in ("x-ui", "nginx"):
                if running[service]:
                    run("systemctl", "stop", service)
            prepare = apply_files(members, staging, Path("/"), rollback, changes)
            if has_nginx:
                # apt may create a default site absent on the source server.
                archived = {m.name for m in members}
                for parent, directories, files in os.walk("/etc/nginx", followlinks=False):
                    for name in files + [d for d in directories if (Path(parent) / d).is_symlink()]:
                        extra = Path(parent) / name
                        if "files" + str(extra) not in archived:
                            prepare(extra)
                            if extra.exists() or extra.is_symlink():
                                extra.unlink()
            destination = Path(db_path)
            prepare(destination)
            shutil.copyfile(staging / "database/x-ui.db", destination)
            destination.chmod(0o600)
            for suffix in ("-wal", "-shm"):
                sidecar = Path(db_path + suffix)
                if sidecar.exists():
                    prepare(sidecar)
                    sidecar.unlink()
            dropin = Path("/etc/systemd/system/x-ui.service.d/30-fork-nginx.conf")
            prepare(dropin)
            dropin.write_text("[Service]\nReadWritePaths=-/etc/nginx -/var/www -/usr/share/nginx/html\n")
            dropin.chmod(0o644)
            run("systemctl", "daemon-reload")
            if has_nginx:
                run("nginx", "-t")
                run("systemctl", "start", "nginx")
            run("systemctl", "start", "x-ui")
            for _ in range(3):
                time.sleep(1)
                run("systemctl", "is-active", "--quiet", "x-ui")
            run("systemctl", "enable", "x-ui")
            if has_nginx:
                run("systemctl", "enable", "nginx")
        except Exception:
            for service in ("x-ui", "nginx"):
                if active(service):
                    run("systemctl", "stop", service)
            undo_files(changes)
            run("systemctl", "daemon-reload")
            for service, was_running in running.items():
                if was_running:
                    run("systemctl", "start", service)
            shutil.rmtree(rollback)
            raise
        # acme.sh keeps its state in the backup, but its crontab is outside it.
        acme = Path("/root/.acme.sh/acme.sh")
        if acme.is_file():
            result = subprocess.run(["sh", str(acme), "--install-cronjob"], check=False)
            if result.returncode:
                print("Warning: certificate renewal cron could not be registered; configure it manually.")
        shutil.rmtree(rollback)
        print("Restored. Login credentials and panel URL are unchanged from the backup.")
        print("Check provider firewall rules, DNS, external storage access and certificate renewal.")
        subprocess.run([executable, "setting", "-show"], check=False,
                       env=dict(os.environ, XUI_DB_FOLDER=str(Path(db_path).parent), XUI_DB_TYPE="sqlite", XUI_DB_DSN=""))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--backup", required=True)
    parser.add_argument("--check", action="store_true")
    parser.add_argument("--replace", action="store_true")
    try:
        restore(parser.parse_args())
    except (ValueError, OSError, tarfile.TarError, sqlite3.Error, subprocess.CalledProcessError) as error:
        print("Restore failed:", error, file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
