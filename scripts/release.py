#!/usr/bin/env python3

import argparse
from http.client import HTTPException
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
from urllib import error, request

from trigger_deployment import NoRedirect


TAG = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)")
SHA = re.compile(r"[a-f0-9]{40}")
DIGEST = re.compile(r"sha256:[a-f0-9]{64}")
COMPONENTS = {
    "backend": ("texinroistot-server", "Dockerfile.prod"),
    "frontend": ("texinroistot-ui", "Dockerfile.prod"),
    "importer": ("texinroistot-server", "Dockerfile.importer"),
    "migrator": ("texinroistot-server", "Dockerfile.migrator"),
}


class ReleaseError(ValueError):
    pass


def github_failure_detail(failure):
    try:
        body = json.loads(failure.read(8192))
    except (ValueError, OSError, HTTPException):
        return "Response details omitted"
    if not isinstance(body, dict):
        return "Response details omitted"
    message = body.get("message")
    known_messages = {
        "Bad credentials", "Not Found", "Validation Failed", "Reference already exists",
        "Reference update failed", "Resource not accessible by integration",
        "Resource not accessible by personal access token", "Internal Server Error",
        "Service Unavailable", "Invalid request", "Problems parsing JSON",
        "Body should be a JSON object",
    }
    detail = message if isinstance(message, str) and message in known_messages else "Response details omitted"
    if isinstance(message, str):
        for prefix, description in (
            ("API rate limit exceeded", "API rate limit exceeded"),
            ("You have exceeded a secondary rate limit", "Secondary API rate limit exceeded"),
            ("Repository rule violations found", "Repository rule violations found"),
            ("refusing to allow a GitHub App to create or update workflow", "Workflow update permission required"),
        ):
            if message.startswith(prefix):
                detail = description
                break
    errors = body.get("errors", [])
    if isinstance(errors, list):
        codes = {item.get("code") for item in errors if isinstance(item, dict) and isinstance(item.get("code"), str)}
        known_codes = codes & {"missing", "missing_field", "invalid", "already_exists", "unprocessable", "custom"}
        if known_codes:
            detail += "; validation codes: " + ", ".join(sorted(known_codes))
    return detail


def run(command, check=True):
    result = subprocess.run(command, text=True, capture_output=True, timeout=1800)
    if check and result.returncode:
        raise ReleaseError(f"{command[0]} operation failed; refusing to publish or overwrite a release")
    return result


def version(tag):
    return tuple(int(part) for part in TAG.fullmatch(tag).groups())


def release_tags():
    tags = run(["git", "tag", "--list", "v*"]).stdout.splitlines()
    return sorted((tag for tag in tags if TAG.fullmatch(tag)), key=version)


def tag_commit(tag):
    return run(["git", "rev-parse", f"refs/tags/{tag}^{{commit}}"]).stdout.strip()


def select_release(source_sha):
    tags = release_tags()
    matching = [tag for tag in tags if tag_commit(tag) == source_sha]
    if len(matching) > 1:
        raise ReleaseError("Multiple release tags point to this commit; operator review required")
    if matching:
        return matching[0]
    if not tags:
        return "v0.0.1"
    latest = tags[-1]
    if run(["git", "merge-base", "--is-ancestor", tag_commit(latest), source_sha], check=False).returncode:
        raise ReleaseError("A newer or divergent release already exists; refusing an out-of-order release")
    major, minor, patch = version(latest)
    return f"v{major}.{minor}.{patch + 1}"


def inspect_image(reference, source_sha):
    result = run(["docker", "buildx", "imagetools", "inspect", reference,
                  "--format", "{{json .Manifest}}"], check=False)
    if result.returncode:
        if result.stderr.strip() == f"ERROR: {reference}: not found":
            return None
        raise ReleaseError(f"Registry inspection failed for {reference}; absence was not established (exit {result.returncode})")
    digest = json.loads(result.stdout).get("digest", "")
    if not DIGEST.fullmatch(digest):
        raise ReleaseError("Registry returned an invalid image digest")
    repository = reference.split("@", 1)[0].split(":", 1)[0]
    result = run(["docker", "buildx", "imagetools", "inspect",
                  f"{repository}@{digest}", "--format", "{{json .Image}}"])
    config = json.loads(result.stdout)
    if "config" not in config:
        config = config.get("linux/amd64", {})
    if config.get("os") != "linux" or config.get("architecture") != "amd64":
        raise ReleaseError(f"Image {reference} is not linux/amd64; refusing overwrite")
    revision = config.get("config", {}).get("Labels", {}).get("org.opencontainers.image.revision")
    if not isinstance(revision, str) or not SHA.fullmatch(revision) or (source_sha is not None and revision != source_sha):
        actual = revision if isinstance(revision, str) and SHA.fullmatch(revision) else "missing or invalid"
        raise ReleaseError(f"Image {reference} source mismatch: expected {source_sha}, found {actual}; refusing overwrite")
    return digest


def ensure_tag(image, tag, digest, source_sha):
    reference = f"{image}:{tag}"
    existing = inspect_image(reference, source_sha)
    if existing is not None and existing != digest:
        raise ReleaseError(f"Immutable image tag {reference} already has a different digest: expected {digest}, found {existing}")
    if existing is None:
        run(["docker", "buildx", "imagetools", "create", "--prefer-index=false",
             "--tag", reference, f"{image}@{digest}"])
    if inspect_image(reference, source_sha) != digest:
        raise ReleaseError("Published image digest verification failed")


def build_candidate(repository, component, source_sha):
    image = f"ghcr.io/{repository.lower()}-{component}"
    context, dockerfile = COMPONENTS[component]
    with tempfile.TemporaryDirectory() as directory:
        metadata = Path(directory) / "metadata.json"
        run(["docker", "buildx", "build", "--platform", "linux/amd64", "--provenance=false",
             "--label", f"org.opencontainers.image.revision={source_sha}",
             "--label", f"org.opencontainers.image.source=https://github.com/{repository}",
             "--output", f"type=image,name={image},push-by-digest=true,name-canonical=true,push=true",
             "--metadata-file", str(metadata), "--file", f"{context}/{dockerfile}", context])
        digest = json.loads(metadata.read_text()).get("containerimage.digest", "")
    if not DIGEST.fullmatch(digest) or inspect_image(f"{image}@{digest}", source_sha) != digest:
        raise ReleaseError("Built image digest verification failed")
    return digest


def publish_image(repository, component, source_sha):
    image = f"ghcr.io/{repository.lower()}-{component}"
    candidate = build_candidate(repository, component, source_sha)
    source_digest = inspect_image(f"{image}:sha-{source_sha}", source_sha)
    digest = source_digest or candidate
    ensure_tag(image, f"sha-{source_sha}", digest, source_sha)


def reserve_release(repository, source_sha):
    candidate = select_release(source_sha)
    tags = release_tags()
    digests = {}
    for component in COMPONENTS:
        image = f"ghcr.io/{repository.lower()}-{component}"
        digest = inspect_image(f"{image}:sha-{source_sha}", source_sha)
        if digest is None:
            raise ReleaseError("All four source images must exist before reserving a version")
        digests[image] = digest
    for attempt in range(100):
        conflicts = []
        for image, digest in digests.items():
            existing = inspect_image(f"{image}:{candidate}", None)
            if existing is not None and existing != digest:
                conflicts.append(image)
        if not conflicts:
            if candidate not in tags:
                github_api(repository, "git/refs", {"ref": f"refs/tags/{candidate}", "sha": source_sha})
            return candidate
        if candidate in tags:
            raise ReleaseError(f"Reserved version {candidate} conflicts with existing images; operator review required")
        print(f"Skipping unreserved image version {candidate}; existing images differ from this commit")
        major, minor, patch = version(candidate)
        candidate = f"v{major}.{minor}.{patch + 1}"
    raise ReleaseError("Too many occupied image versions; operator review required")


def require_reservation(release_tag, source_sha):
    if release_tag not in release_tags() or tag_commit(release_tag) != source_sha:
        raise ReleaseError("The release version must be reserved for this commit before promotion or publication")


def promote_images(repository, release_tag, source_sha):
    require_reservation(release_tag, source_sha)
    digests = {}
    for component in COMPONENTS:
        image = f"ghcr.io/{repository.lower()}-{component}"
        try:
            digest = inspect_image(f"{image}:sha-{source_sha}", source_sha)
            if digest is None:
                raise ReleaseError("Source image is missing; all four must exist before promotion")
            existing = inspect_image(f"{image}:{release_tag}", source_sha)
            if existing is not None and existing != digest:
                raise ReleaseError(f"Release and source image digests differ: source {digest}, release {existing}")
            digests[image] = digest
        except ReleaseError as failure:
            raise ReleaseError(f"Promotion preflight for {component} at {release_tag}: {failure}") from None
    for image, digest in digests.items():
        ensure_tag(image, release_tag, digest, source_sha)


def github_api(repository, path, payload=None, missing_ok=False):
    token = os.environ.get("GH_TOKEN", "")
    if not token:
        raise ReleaseError("The release job requires its scoped GitHub token")
    operation = request.Request(
        f"https://api.github.com/repos/{repository}/{path}",
        data=json.dumps(payload).encode() if payload is not None else None,
        headers={"Authorization": f"Bearer {token}", "Accept": "application/vnd.github+json",
                 "Content-Type": "application/json", "X-GitHub-Api-Version": "2022-11-28"},
    )
    try:
        with request.build_opener(NoRedirect()).open(operation, timeout=30) as response:
            return json.load(response)
    except error.HTTPError as failure:
        status = failure.code
        try:
            if status == 404 and missing_ok:
                return None
            detail = github_failure_detail(failure)
        finally:
            failure.close()
        operation_name = {
            "git/refs": "create Git tag",
            "releases/generate-notes": "generate release notes",
            "releases": "create GitHub Release",
        }.get(path, "look up GitHub Release")
        raise ReleaseError(f"GitHub API failed during {operation_name} (HTTP {status}): {detail}") from None


def publish_release(repository, release_tag, source_sha):
    require_reservation(release_tag, source_sha)
    tags = release_tags()
    digests = {}
    for component in COMPONENTS:
        image = f"ghcr.io/{repository.lower()}-{component}"
        digest = inspect_image(f"{image}:{release_tag}", source_sha)
        if digest is None or inspect_image(f"{image}:sha-{source_sha}", source_sha) != digest:
            raise ReleaseError("All four matching release/source images must exist before release creation")
        digests[image] = digest
    existing = github_api(repository, f"releases/tags/{release_tag}", missing_ok=True)
    if existing is not None:
        if existing.get("tag_name") != release_tag or existing.get("draft") or existing.get("prerelease"):
            raise ReleaseError("An incompatible GitHub Release already exists")
    else:
        previous = [tag for tag in tags if version(tag) < version(release_tag)]
        payload = {"tag_name": release_tag, "target_commitish": source_sha}
        for tag in reversed(previous):
            published = github_api(repository, f"releases/tags/{tag}", missing_ok=True)
            if published is not None and not published.get("draft") and not published.get("prerelease"):
                payload["previous_tag_name"] = tag
                break
        notes = github_api(repository, "releases/generate-notes", payload)
        body = notes["body"] + f"\n\nSource commit: `{source_sha}`\n\nImage digests:\n"
        body += "\n".join(f"- `{image}@{digest}`" for image, digest in digests.items())
        github_api(repository, "releases", {
            "tag_name": release_tag, "target_commitish": source_sha, "name": release_tag,
            "body": body, "draft": False, "prerelease": False,
            "make_latest": "true" if not tags or version(release_tag) >= version(tags[-1]) else "false",
        })
    return not tags or version(release_tag) >= version(tags[-1])


def handoff_eligible(release_tag, source_sha):
    tags = release_tags()
    return bool(tags) and tags[-1] == release_tag and tag_commit(release_tag) == source_sha


def main():
    parser = argparse.ArgumentParser(description="Publish immutable application releases from trusted CI")
    parser.add_argument("operation", choices=["plan", "image", "reserve", "promote", "publish", "handoff-check"])
    parser.add_argument("--source-sha", required=True)
    parser.add_argument("--repository", required=True)
    parser.add_argument("--release-tag")
    parser.add_argument("--component", choices=COMPONENTS)
    args = parser.parse_args()
    try:
        if not SHA.fullmatch(args.source_sha) or not re.fullmatch(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+", args.repository):
            raise ReleaseError("Invalid release source or repository")
        if run(["git", "rev-parse", "HEAD"]).stdout.strip() != args.source_sha:
            raise ReleaseError("Checkout must match the tested source commit")
        if args.operation not in ("plan", "image", "reserve") and not TAG.fullmatch(args.release_tag or ""):
            raise ReleaseError("A strict release tag is required")
        if args.operation == "plan":
            output = f"release_tag={select_release(args.source_sha)}"
        elif args.operation == "reserve":
            output = f"release_tag={reserve_release(args.repository, args.source_sha)}"
        elif args.operation == "image":
            if not args.component:
                raise ReleaseError("An image component is required")
            publish_image(args.repository, args.component, args.source_sha)
            output = f"component={args.component}"
        elif args.operation == "promote":
            promote_images(args.repository, args.release_tag, args.source_sha)
            output = f"release_tag={args.release_tag}"
        elif args.operation == "publish":
            eligible = publish_release(args.repository, args.release_tag, args.source_sha)
            output = f"handoff={str(eligible).lower()}"
        else:
            output = f"handoff={str(handoff_eligible(args.release_tag, args.source_sha)).lower()}"
        if os.environ.get("GITHUB_OUTPUT"):
            with Path(os.environ["GITHUB_OUTPUT"]).open("a") as destination:
                destination.write(output + "\n")
        print(output)
        return 0
    except ReleaseError as failure:
        print(f"Release {args.operation} failed: {failure}", file=sys.stderr)
        return 1
    except (ValueError, KeyError, TypeError, OSError, error.URLError, HTTPException, subprocess.TimeoutExpired) as failure:
        print(f"Release {args.operation} failed ({type(failure).__name__}); details omitted. Review the release state before retrying.", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
