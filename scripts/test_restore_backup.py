import importlib.util
from contextlib import closing
import io
import json
import os
import sqlite3
import tarfile
import tempfile
import unittest
from unittest import mock
from pathlib import Path

spec = importlib.util.spec_from_file_location("restore_backup", Path(__file__).with_name("restore_backup.py"))
restore = importlib.util.module_from_spec(spec)
spec.loader.exec_module(restore)


class RestoreBackupTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        db_path = self.root / "snapshot.db"
        with closing(sqlite3.connect(db_path)) as db:
            db.execute("CREATE TABLE test (value TEXT)")
            db.execute("INSERT INTO test VALUES ('restored')")
            db.commit()
        self.snapshot = db_path.read_bytes()

    def archive(self, extras=(), manifest=None, snapshot=None):
        data = io.BytesIO()
        with tarfile.open(fileobj=data, mode="w:gz") as archive:
            files = {
                "manifest.json": json.dumps(manifest or {"version": "3.10.2"}).encode(),
                "database/x-ui.db": self.snapshot if snapshot is None else snapshot,
                "files/usr/local/x-ui/x-ui": b"panel binary",
                "files/etc/systemd/system/x-ui.service": b"service",
                "files/etc/nginx/nginx.conf": b"events {} http {}",
                "files/var/www/site/index.html": b"restored website",
            }
            for name, value in files.items():
                member = tarfile.TarInfo(name)
                member.size = len(value)
                member.mode = 0o600
                archive.addfile(member, io.BytesIO(value))
            for member in extras:
                archive.addfile(member)
        data.seek(0)
        return tarfile.open(fileobj=data, mode="r:gz")

    def inspect(self, archive):
        staging = self.root / "staging"
        staging.mkdir(exist_ok=True)
        with archive:
            return restore.inspect_archive(archive, staging)

    def test_accepts_legacy_snapshot_and_manifest(self):
        manifest, members, executable, db = self.inspect(self.archive())
        self.assertEqual(manifest["version"], "3.10.2")
        self.assertEqual(executable, "/usr/local/x-ui/x-ui")
        self.assertEqual(db, "/etc/x-ui/x-ui.db")
        self.assertEqual(len(members), 6)

    def test_rejects_traversal_duplicates_hardlinks_and_symlink_children(self):
        hardlink = tarfile.TarInfo("files/link")
        hardlink.type = tarfile.LNKTYPE
        hardlink.linkname = "files/usr/local/x-ui/x-ui"
        link = tarfile.TarInfo("files/var/www/site")
        link.type = tarfile.SYMTYPE
        link.linkname = "/etc"
        for extras in ([tarfile.TarInfo("files/../../escape")],
                       [tarfile.TarInfo("files//etc/escape")],
                       [tarfile.TarInfo("manifest.json")], [hardlink], [link]):
            with self.subTest(extras=extras), self.assertRaises(ValueError):
                self.inspect(self.archive(extras))
        self.assertFalse((self.root / "staging/manifest.json").exists())

    def test_rejects_windows_unknown_format_and_corrupt_database(self):
        for manifest in ({"version": "3", "os": "windows"},
                         {"version": "3", "formatVersion": 2},
                         {"version": "3", "dbPath": "/etc/../passwd"}):
            with self.subTest(manifest=manifest), self.assertRaises(ValueError):
                self.inspect(self.archive(manifest=manifest))
        with self.assertRaises(sqlite3.Error):
            self.inspect(self.archive(snapshot=b"not a sqlite database"))

    def test_files_and_symlinks_can_be_rolled_back(self):
        destination = self.root / "target"
        destination.mkdir()
        existing = destination / "var/www/site/index.html"
        existing.parent.mkdir(parents=True)
        existing.write_text("old website")
        link = tarfile.TarInfo("files/etc/nginx/sites-enabled/site")
        link.type = tarfile.SYMTYPE
        link.linkname = "/etc/nginx/sites-available/site"
        if os.name == "nt":
            # Windows symlink creation requires privileges; file rollback is still tested.
            extras = []
        else:
            extras = [link]
        _, members, _, _ = self.inspect(self.archive(extras))
        rollback = self.root / "rollback"
        rollback.mkdir()
        changes = []
        restore.apply_files(members, self.root / "staging", destination, rollback, changes)
        self.assertEqual(existing.read_text(), "restored website")
        self.assertTrue((destination / "usr/local/x-ui/x-ui").exists())
        if extras:
            self.assertEqual(os.readlink(destination / "etc/nginx/sites-enabled/site"), link.linkname)
        restore.undo_files(changes)
        self.assertEqual(existing.read_text(), "old website")
        self.assertFalse((destination / "usr").exists())

    def test_check_mode_never_installs_packages_or_changes_services(self):
        archive_path = self.root / "full.tar.gz"
        with self.archive() as archive:
            archive_path.write_bytes(archive.fileobj.fileobj.getvalue())
        args = type("Args", (), {"backup": str(archive_path), "check": True, "replace": False})()
        with mock.patch.object(restore, "check_platform"), mock.patch.object(restore.subprocess, "run") as command:
            restore.restore(args)
        command.assert_not_called()

    def test_partial_write_failure_restores_original_binary(self):
        destination = self.root / "target"
        old_binary = destination / "usr/local/x-ui/x-ui"
        old_binary.parent.mkdir(parents=True)
        old_binary.write_text("original binary")
        _, members, _, _ = self.inspect(self.archive())
        rollback = self.root / "rollback"
        rollback.mkdir()
        changes = []
        copyfile = restore.shutil.copyfile

        def fail_nginx(source, target, **kwargs):
            if str(target).endswith("nginx.conf"):
                raise OSError("simulated disk write failure")
            return copyfile(source, target, **kwargs)

        with mock.patch.object(restore.shutil, "copyfile", side_effect=fail_nginx), self.assertRaises(OSError):
            restore.apply_files(members, self.root / "staging", destination, rollback, changes)
        self.assertTrue((rollback / "paths.json").exists())
        restore.undo_files(changes)
        self.assertEqual(old_binary.read_text(), "original binary")
        self.assertFalse((destination / "etc").exists())

    def test_checks_distribution_and_actual_binary_architecture(self):
        binary = self.root / "panel"
        header = bytearray(20)
        header[:4] = b"\x7fELF"
        header[5] = 1
        header[18:20] = (62).to_bytes(2, "little")
        binary.write_bytes(header)
        manifest = {"osRelease": 'ID=ubuntu\nVERSION_ID="24.04"\n'}
        with mock.patch.object(restore.platform, "system", return_value="Linux"), \
             mock.patch.object(restore.platform, "machine", return_value="x86_64"), \
             mock.patch.object(Path, "read_text", return_value=manifest["osRelease"]), \
             mock.patch.object(Path, "is_dir", return_value=True):
            restore.check_platform(manifest, binary)
            with self.assertRaises(ValueError):
                restore.check_platform({"osRelease": "ID=debian\nVERSION_ID=12"}, binary)
            header[18:20] = (183).to_bytes(2, "little")
            binary.write_bytes(header)
            with self.assertRaises(ValueError):
                restore.check_platform(manifest, binary)


if __name__ == "__main__":
    unittest.main()
