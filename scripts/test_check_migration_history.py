import contextlib
import io
from pathlib import Path
import subprocess
import tempfile
import unittest

from check_migration_history import MIGRATIONS, check_history


class MigrationHistoryTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        (self.root / MIGRATIONS).mkdir(parents=True)
        self.git("init", "--quiet")
        self.add_pair("000002_initial")
        self.git("add", ".")
        self.git("-c", "user.name=Migration Test", "-c", "user.email=test@example.invalid",
                 "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "Base migrations")

    def git(self, *args):
        return subprocess.check_output(["git", *args], cwd=self.root, stderr=subprocess.STDOUT)

    def add_pair(self, name):
        for direction in ("up", "down"):
            (self.root / MIGRATIONS / f"{name}.{direction}.sql").write_text("SELECT 1;\n")

    def check(self):
        with contextlib.redirect_stdout(io.StringIO()):
            check_history(self.root, "HEAD")

    def test_allows_unchanged_history_and_newer_pair(self):
        self.check()
        self.add_pair("000003_next")
        self.check()

    def test_rejects_edit_of_merged_migration(self):
        (self.root / MIGRATIONS / "000002_initial.up.sql").write_text("SELECT 2;\n")
        with self.assertRaisesRegex(ValueError, "append-only"):
            self.check()

    def test_rejects_deletion_of_merged_pair(self):
        self.add_pair("000003_next")
        for path in (self.root / MIGRATIONS).glob("000002_*"):
            path.unlink()
        with self.assertRaisesRegex(ValueError, "append-only"):
            self.check()

    def test_rejects_renumbering(self):
        for path in (self.root / MIGRATIONS).glob("000002_*"):
            path.rename(path.with_name(path.name.replace("000002", "000003")))
        with self.assertRaisesRegex(ValueError, "append-only"):
            self.check()

    def test_rejects_older_new_migration(self):
        self.add_pair("000001_older")
        with self.assertRaisesRegex(ValueError, "newer than"):
            self.check()

    def test_rejects_missing_reverse_migration(self):
        (self.root / MIGRATIONS / "000003_next.up.sql").write_text("SELECT 1;\n")
        with self.assertRaisesRegex(ValueError, "matching nonempty"):
            self.check()

    def test_rejects_duplicate_version(self):
        self.add_pair("000002_duplicate")
        with self.assertRaisesRegex(ValueError, "Duplicate"):
            self.check()

    def test_rejects_unknown_base(self):
        with self.assertRaises(subprocess.CalledProcessError):
            check_history(self.root, "missing-base-ref")


if __name__ == "__main__":
    unittest.main()
