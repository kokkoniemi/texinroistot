# Releases and retries

## Publication

Pushes to `main` and manual CI runs on `main` publish releases after application and migration checks pass. Pull requests run checks only.

`.github/workflows/ci.yml` calls `images.yml` to:
1. Build backend, frontend, importer, and migrator images as `ghcr.io/<owner>/<repo>-<component>:sha-<full-commit-sha>`.
2. Reserve a `vX.Y.Z` Git tag for the tested commit.
3. Tag the same image digests with that version and verify all four.
4. Create the GitHub Release with its source SHA and image digests.

No branch, `latest`, or semver-alias image tags are published. CI serializes runs per branch without canceling an active release.

## Failed releases

- Retry the failed run for the same commit. Existing reservations and image digests are reused.
- A Git tag or image alone does not mean publication finished; check for the completed GitHub Release.
- Never move tags or overwrite images to fix a failed release. Conflicting image-only versions are skipped.
- Retrying a completed release creates no new version. Finishing an older release does not mark it latest.
- Rerunning an old workflow uses its old scripts. To use a workflow fix, run CI on the corrected commit.

## Permissions

Jobs declare their permissions: image publication needs package writes; tag and release creation need repository writes. Checks use read-only repository access. Protect release tags and restrict package writes to trusted publishers. Review visibility for all four GHCR packages.

## Local checks

```bash
python3 -B -m unittest discover -s scripts -p 'test_*.py'
```

These tests use temporary repositories and mocked APIs; they do not publish releases.
