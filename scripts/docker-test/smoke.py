#!/usr/bin/env python3
"""Exercise only the disposable compose.test.yml environment; standard library only."""
from contextlib import closing
import hashlib
import http.cookiejar
import io
import json
from pathlib import Path
import sqlite3
import subprocess
import tarfile
import tempfile
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
COMPOSE = ["docker", "compose", "-p", "fork-backup-test", "-f", str(ROOT / "compose.test.yml")]
BASE = "http://127.0.0.1:12053"
API = "/panel/api/maintenance"


def docker(*args):
    return subprocess.check_output(COMPOSE + list(args), cwd=ROOT, text=True).strip()


class Client:
    def __init__(self):
        self.http = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
        self.csrf = self.request("/csrf-token")["obj"]
        self.request("/login", {"username": "dev-admin", "password": "LocalTest-3xui-2026"})
        self.csrf = self.request("/csrf-token")["obj"]

    def request(self, path, data=None, success=True, raw=False):
        headers = {"X-CSRF-Token": getattr(self, "csrf", "")}
        payload = None
        if data is not None:
            payload = json.dumps(data).encode()
            headers["Content-Type"] = "application/json"
        with self.http.open(urllib.request.Request(BASE + path, payload, headers), timeout=180) as response:
            body = response.read()
        if raw:
            return body
        value = json.loads(body)
        if value["success"] != success:
            raise AssertionError(f"{path}: {value}")
        return value

    def backups(self):
        return self.request(API + "/backups")["obj"]


def check_website(expected):
    for attempt in range(20):
        req = urllib.request.Request("http://127.0.0.1:18080/", headers={"Host": "backup-test.local"})
        with urllib.request.urlopen(req, timeout=10) as response:
            body = response.read().decode()
        if expected in body:
            return
        time.sleep(0.1)
    raise AssertionError(f"Website did not contain {expected!r}: {body}")


def ready_client():
    for attempt in range(30):
        try:
            return Client()
        except (urllib.error.URLError, ConnectionError):
            if attempt == 29:
                raise
            time.sleep(1)


def main():
    client = ready_client()
    settings = client.backups()["settings"]
    settings.update(enabled=False, intervalHours=24, keep=2, remoteHost="")
    client.request(API + "/backups/settings", settings)
    print("PASS: authenticated settings API", flush=True)

    site = {"configPath": "/etc/nginx/conf.d/backup-test.conf", "root": "/var/www/backup-test",
            "config": "server { listen 80; server_name backup-test.local; root /var/www/backup-test; index index.html; }\n",
            "html": "<h1>Restorable test website</h1>\n"}
    client.request(API + "/nginx", site)
    check_website("Restorable test website")
    client.request(API + "/nginx", {**site, "config": "this is invalid nginx;", "html": "BROKEN"}, success=False)
    check_website("Restorable test website")
    print("PASS: nginx apply, live HTTP and rollback on invalid configuration", flush=True)

    client.request(API + "/backups/run", {})
    name = client.backups()["status"]["file"]
    archive_bytes = client.request(API + "/backups/download/" + name, raw=True)
    with tarfile.open(fileobj=io.BytesIO(archive_bytes), mode="r:gz") as archive:
        names = set(archive.getnames())
        for required in ["database/x-ui.db", "files/app/x-ui", "files/etc/nginx/nginx.conf",
                         "files/etc/nginx/conf.d/backup-test.conf", "files/var/www/backup-test/index.html", "manifest.json"]:
            assert required in names, required
        assert not any("full-backups/" in n or n.endswith(("x-ui.db-wal", "x-ui.db-shm")) for n in names)
        assert archive.extractfile("files/var/www/backup-test/index.html").read().decode() == site["html"]
        manifest = json.load(archive.extractfile("manifest.json"))
        assert manifest["os"] == "linux" and manifest["dbType"] == "sqlite"
        with tempfile.TemporaryDirectory() as temp:
            db = Path(temp) / "snapshot.db"
            db.write_bytes(archive.extractfile("database/x-ui.db").read())
            with closing(sqlite3.connect(db)) as connection:
                assert connection.execute("PRAGMA quick_check").fetchone()[0] == "ok"
                assert connection.execute("SELECT COUNT(*) FROM users").fetchone()[0] > 0
    archive_path = ROOT / ".cache" / "docker-test-last.tar.gz"
    archive_path.parent.mkdir(exist_ok=True)
    archive_path.write_bytes(archive_bytes)
    assert docker("exec", "-T", "panel", "stat", "-c", "%a", "/etc/x-ui/full-backups/" + name) == "600"
    print("PASS: real downloaded archive, SQLite integrity, binary, nginx, website and private permissions", flush=True)

    settings.update(remoteHost="backup-target", remoteUser="backup", remotePort=22,
                    remotePath="/srv/backups/3x-ui", identityFile="/root/.ssh/backup_ed25519")
    client.request(API + "/backups/settings", settings)
    client.request(API + "/backups/run", {})
    name = client.backups()["status"]["file"]
    downloaded = client.request(API + "/backups/download/" + name, raw=True)
    remote_hash = docker("exec", "-T", "backup-target", "sha256sum", "/srv/backups/3x-ui/" + name).split()[0]
    assert remote_hash == hashlib.sha256(downloaded).hexdigest()
    assert not docker("exec", "-T", "backup-target", "find", "/srv/backups/3x-ui", "-name", "*.partial")
    print("PASS: SSH/SFTP delivery, identical SHA256 and atomic final filename", flush=True)

    bad_settings = {**settings, "remotePath": "/srv/missing-directory"}
    client.request(API + "/backups/settings", bad_settings)
    client.request(API + "/backups/run", {}, success=False)
    failed = client.backups()
    assert failed["status"]["error"] and len(failed["files"]) == 2
    assert any(f["name"] == failed["status"]["file"] for f in failed["files"])
    print("PASS: failed upload is visible, local archive retained, retention enforced", flush=True)

    client.request(API + "/backups/settings", settings)
    docker("restart", "panel")
    client = ready_client()
    assert client.backups()["settings"] == settings
    check_website("Restorable test website")
    print("PASS: settings, archives and website survive container restart", flush=True)

    settings["enabled"] = True
    client.request(API + "/backups/settings", settings)
    # Make the schedule due in this isolated test DB without waiting 24 hours.
    docker("exec", "-T", "panel", "python3", "-c",
           "import sqlite3,json; c=sqlite3.connect('/etc/x-ui/x-ui.db'); "
           "s=json.loads(c.execute(\"SELECT value FROM settings WHERE key='fork.fullBackupStatus'\").fetchone()[0]); "
           "s['lastAttempt']='2000-01-01T00:00:00Z'; "
           "c.execute(\"UPDATE settings SET value=? WHERE key='fork.fullBackupStatus'\",(json.dumps(s),)); c.commit(); c.close()")
    print("Waiting for the real scheduler (up to 75 seconds)...", flush=True)
    for attempt in range(75):
        status = client.backups()["status"]
        if status["file"] != failed["status"]["file"] and not status["error"]:
            break
        time.sleep(1)
    else:
        raise AssertionError("Scheduled backup did not complete")
    assert docker("exec", "-T", "backup-target", "test", "-f", "/srv/backups/3x-ui/" + status["file"]) == ""
    settings["enabled"] = False
    client.request(API + "/backups/settings", settings)
    print("PASS: actual scheduled backup and remote delivery", flush=True)
    print("ALL CHECKS PASSED. Panel: " + BASE + "/panel/maintenance", flush=True)


if __name__ == "__main__":
    main()
