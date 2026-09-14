#!/usr/bin/env python3

import argparse
from pathlib import Path
import re
import subprocess
import sys


MIGRATIONS = Path("texinroistot-server/internal/db/migrations")
NAME = re.compile(r"([0-9]{6})_([a-z][a-z0-9_]+)\.(up|down)\.sql")


def check_history(root, base_ref):
    migrations = root / MIGRATIONS
    pairs = {}
    for path in sorted(migrations.iterdir()):
        match = NAME.fullmatch(path.name)
        if not match or not path.is_file() or path.is_symlink():
            raise ValueError(f"Invalid migration file: {path.name}")
        version = int(match[1])
        if version == 0:
            raise ValueError("Migration versions must start at 1")
        name, directions = pairs.setdefault(version, (match[2], set()))
        if name != match[2] or match[3] in directions:
            raise ValueError(f"Duplicate migration version: {version}")
        if not path.read_text().strip():
            raise ValueError(f"Empty migration: {path.name}")
        directions.add(match[3])

    if not pairs or any(directions != {"up", "down"} for _, directions in pairs.values()):
        raise ValueError("Every migration must have matching nonempty up and down SQL")

    existing = subprocess.check_output(
        ["git", "ls-tree", "-r", "--name-only", base_ref, "--", str(MIGRATIONS)],
        cwd=root, text=True,
    ).splitlines()
    highest = 0
    for filename in existing:
        match = NAME.fullmatch(Path(filename).name)
        if not match:
            continue
        highest = max(highest, int(match[1]))
        original = subprocess.check_output(["git", "show", f"{base_ref}:{filename}"], cwd=root)
        path = root / filename
        if not path.is_file() or path.is_symlink() or path.read_bytes() != original:
            raise ValueError(f"Merged migrations are append-only: {filename}")

    for path in migrations.iterdir():
        if str(path.relative_to(root)) not in existing and int(NAME.fullmatch(path.name)[1]) <= highest:
            raise ValueError(f"New migrations must be newer than version {highest}: {path.name}")

    print(f"Migration history valid: {len(pairs)} paired versions; base {base_ref}")


def main():
    parser = argparse.ArgumentParser(description="Check migration names, pairs, and append-only history")
    parser.add_argument("--base-ref", default="HEAD")
    args = parser.parse_args()
    try:
        check_history(Path(__file__).resolve().parent.parent, args.base_ref)
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        print(error, file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
