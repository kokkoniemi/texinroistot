import contextlib
import io
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
from urllib import error

import release


SOURCE = "a" * 40
DIGEST = "sha256:" + "b" * 64
OTHER_DIGEST = "sha256:" + "c" * 64
REPOSITORY = "example/comics"
IMAGE = f"ghcr.io/{REPOSITORY}-backend"


class VersionTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        original = Path.cwd()
        os.chdir(temporary.name)
        self.addCleanup(os.chdir, original)
        environment = patch.dict(os.environ, {
            "PATH": os.defpath, "HOME": temporary.name,
            "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": "/dev/null",
        }, clear=True)
        environment.start()
        self.addCleanup(environment.stop)
        release.run(["git", "init", "--quiet"])
        self.first = self.commit()

    def commit(self):
        release.run(["git", "-c", "user.name=Synthetic Test", "-c", "user.email=test@example.invalid",
                     "commit", "--quiet", "--allow-empty", "-m", "synthetic"])
        return release.run(["git", "rev-parse", "HEAD"]).stdout.strip()

    def tag(self, tag, annotated=False):
        command = ["git", "-c", "user.name=Synthetic Test", "-c", "user.email=test@example.invalid", "tag"]
        if annotated:
            command += ["-a", "-m", "synthetic"]
        release.run(command + [tag])

    def test_first_release_and_strict_numeric_patch_increment(self):
        self.assertEqual(release.select_release(self.first), "v0.0.1")
        for tag in ("v0.0.9", "v0.0.10", "v9.0.0-rc1", "v01.0.0", "v2.0"):
            self.tag(tag)
        self.assertEqual(release.select_release(self.commit()), "v0.0.11")

    def test_reuses_annotated_tag_on_retry(self):
        self.tag("v0.0.8", annotated=True)
        self.assertEqual(release.select_release(self.first), "v0.0.8")
        self.assertEqual(release.select_release(self.first), "v0.0.8")

    def test_successive_commits_get_distinct_versions(self):
        self.tag("v0.0.8")
        second = self.commit()
        self.assertEqual(release.select_release(second), "v0.0.9")
        self.tag("v0.0.9")
        self.assertEqual(release.select_release(self.commit()), "v0.0.10")

    def test_refuses_ambiguous_versions_for_same_commit(self):
        self.tag("v0.0.8")
        self.tag("v0.0.9")
        with self.assertRaises(ValueError):
            release.select_release(self.first)

    def test_old_untagged_commit_cannot_get_a_new_release(self):
        self.commit()
        self.tag("v0.0.9")
        with self.assertRaises(ValueError):
            release.select_release(self.first)

    def test_old_release_retry_cannot_trigger_a_rollback(self):
        self.tag("v0.0.8")
        latest = self.commit()
        self.tag("v0.0.9")
        self.assertEqual(release.select_release(self.first), "v0.0.8")
        self.assertFalse(release.handoff_eligible("v0.0.8", self.first))
        self.assertFalse(release.handoff_eligible("v0.0.9", self.first))
        self.assertTrue(release.handoff_eligible("v0.0.9", latest))

    def test_reservation_requires_a_tag_for_the_exact_commit(self):
        with self.assertRaises(release.ReleaseError):
            release.require_reservation("v0.0.13", self.first)
        self.tag("v0.0.13")
        release.require_reservation("v0.0.13", self.first)
        with self.assertRaises(release.ReleaseError):
            release.require_reservation("v0.0.13", self.commit())

    def test_failed_reservation_is_not_reused_by_next_commit(self):
        self.tag("v0.0.12")
        first_merge = self.commit()

        def reserve(repository, path, payload):
            self.assertEqual(path, "git/refs")
            release.run(["git", "tag", payload["ref"].removeprefix("refs/tags/"), payload["sha"]])

        def inspect(reference, source):
            return DIGEST if ":sha-" in reference else None

        with patch("release.inspect_image", side_effect=inspect), patch("release.github_api", side_effect=reserve):
            self.assertEqual(release.reserve_release(REPOSITORY, first_merge), "v0.0.13")
            second_merge = self.commit()
            self.assertEqual(release.reserve_release(REPOSITORY, second_merge), "v0.0.14")
            self.assertEqual(release.reserve_release(REPOSITORY, first_merge), "v0.0.13")


class ReservationTests(unittest.TestCase):
    def test_skips_legacy_image_only_version_without_overwrites(self):
        with patch("release.select_release", return_value="v0.0.13"), \
                patch("release.release_tags", return_value=["v0.0.12"]), \
                patch("release.inspect_image", side_effect=[DIGEST] * 4 + [OTHER_DIGEST, None, None, None] + [None] * 4), \
                patch("release.github_api") as api, patch("release.ensure_tag") as promote, \
                contextlib.redirect_stdout(io.StringIO()) as output:
            self.assertEqual(release.reserve_release(REPOSITORY, SOURCE), "v0.0.14")
        api.assert_called_once_with(REPOSITORY, "git/refs", {"ref": "refs/tags/v0.0.14", "sha": SOURCE})
        promote.assert_not_called()
        self.assertIn("Skipping unreserved image version v0.0.13", output.getvalue())

    def test_reuses_same_commit_legacy_partial_image_version(self):
        with patch("release.select_release", return_value="v0.0.13"), \
                patch("release.release_tags", return_value=["v0.0.12"]), \
                patch("release.inspect_image", side_effect=[DIGEST] * 4 + [DIGEST, None, DIGEST, None]), \
                patch("release.github_api") as api:
            self.assertEqual(release.reserve_release(REPOSITORY, SOURCE), "v0.0.13")
        api.assert_called_once_with(REPOSITORY, "git/refs", {"ref": "refs/tags/v0.0.13", "sha": SOURCE})

    def test_reserved_tag_retry_does_not_create_another_tag(self):
        with patch("release.select_release", return_value="v0.0.13"), \
                patch("release.release_tags", return_value=["v0.0.12", "v0.0.13", "v0.0.14"]), \
                patch("release.inspect_image", side_effect=[DIGEST] * 4 + [DIGEST, None, None, None]), \
                patch("release.github_api") as api:
            self.assertEqual(release.reserve_release(REPOSITORY, SOURCE), "v0.0.13")
        api.assert_not_called()

    def test_reserved_tag_conflict_fails_without_reallocation(self):
        with patch("release.select_release", return_value="v0.0.13"), \
                patch("release.release_tags", return_value=["v0.0.13"]), \
                patch("release.inspect_image", side_effect=[DIGEST] * 4 + [OTHER_DIGEST] * 4), \
                patch("release.github_api") as api, self.assertRaises(release.ReleaseError):
            release.reserve_release(REPOSITORY, SOURCE)
        api.assert_not_called()

    def test_registry_failure_or_missing_source_never_reserves_a_version(self):
        for result in (None, release.ReleaseError("Registry unavailable")):
            with self.subTest(result=result), patch("release.select_release", return_value="v0.0.13"), \
                    patch("release.release_tags", return_value=["v0.0.12"]), \
                    patch("release.inspect_image", side_effect=[result]), \
                    patch("release.github_api") as api, self.assertRaises(release.ReleaseError):
                release.reserve_release(REPOSITORY, SOURCE)
            api.assert_not_called()

    def test_version_lookup_errors_are_not_treated_as_unoccupied_tags(self):
        with patch("release.select_release", return_value="v0.0.13"), \
                patch("release.release_tags", return_value=["v0.0.12"]), \
                patch("release.inspect_image", side_effect=[DIGEST] * 4 + [release.ReleaseError("Registry unavailable")]), \
                patch("release.github_api") as api, self.assertRaises(release.ReleaseError):
            release.reserve_release(REPOSITORY, SOURCE)
        api.assert_not_called()

    def test_failed_tag_reservation_does_not_promote_images(self):
        with patch("release.select_release", return_value="v0.0.13"), \
                patch("release.release_tags", return_value=["v0.0.12"]), \
                patch("release.inspect_image", side_effect=[DIGEST] * 4 + [None] * 4), \
                patch("release.github_api", side_effect=release.ReleaseError("Tag creation failed")), \
                patch("release.ensure_tag") as promote, self.assertRaises(release.ReleaseError):
            release.reserve_release(REPOSITORY, SOURCE)
        promote.assert_not_called()


class ImageTests(unittest.TestCase):
    def setUp(self):
        reservation = patch("release.require_reservation")
        self.reservation = reservation.start()
        self.addCleanup(reservation.stop)

    def result(self, stdout="", stderr="", status=0):
        return subprocess.CompletedProcess([], status, stdout, stderr)

    def test_registry_absence_must_be_explicit(self):
        reference = f"{IMAGE}:v0.0.9"
        with patch("release.run", return_value=self.result(stderr=f"ERROR: {reference}: not found\n", status=1)):
            self.assertIsNone(release.inspect_image(reference, SOURCE))
        for diagnostic in ("unauthorized", "timeout", "connection refused", "404 from proxy", "not found"):
            with self.subTest(diagnostic=diagnostic), \
                    patch("release.run", return_value=self.result(stderr=diagnostic, status=1)), \
                    self.assertRaises(ValueError):
                release.inspect_image(reference, SOURCE)

    def test_checks_config_at_digest_and_rejects_wrong_revision(self):
        config = {"architecture": "amd64", "os": "linux", "config": {
            "Labels": {"org.opencontainers.image.revision": SOURCE},
        }}
        with patch("release.run", side_effect=[self.result(json.dumps({"digest": DIGEST})),
                                               self.result(json.dumps(config))]) as commands:
            self.assertEqual(release.inspect_image(f"{IMAGE}:v0.0.9", SOURCE), DIGEST)
            self.assertIn(f"{IMAGE}@{DIGEST}", commands.call_args.args[0])
        config["config"]["Labels"]["org.opencontainers.image.revision"] = "d" * 40
        with patch("release.run", side_effect=[self.result(json.dumps({"digest": DIGEST})),
                                               self.result(json.dumps(config))]), self.assertRaises(ValueError):
            release.inspect_image(f"{IMAGE}:v0.0.9", SOURCE)

    def test_retry_preserves_existing_tags_even_if_candidate_digest_changes(self):
        with patch("release.build_candidate", return_value=OTHER_DIGEST), \
                patch("release.inspect_image", return_value=DIGEST), patch("release.run") as commands:
            release.publish_image(REPOSITORY, "backend", SOURCE)
        commands.assert_not_called()

    def test_conflicting_source_and_version_tags_are_never_overwritten(self):
        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.inspect_image", side_effect=[DIGEST, OTHER_DIGEST]), \
                patch("release.run") as commands, self.assertRaises(ValueError):
            release.promote_images(REPOSITORY, "v0.0.9", SOURCE)
        commands.assert_not_called()

    def test_partial_image_push_reuses_source_digest(self):
        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.inspect_image", side_effect=[DIGEST, DIGEST, DIGEST, None, DIGEST, None, DIGEST, None]), \
                patch("release.ensure_tag") as promote:
            release.promote_images(REPOSITORY, "v0.0.9", SOURCE)
        self.assertEqual(promote.call_count, 4)
        promote.assert_any_call(IMAGE, "v0.0.9", DIGEST, SOURCE)

    def test_partial_version_collision_cannot_mix_commits_across_components(self):
        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.inspect_image", side_effect=[DIGEST, None, DIGEST, None, DIGEST, OTHER_DIGEST]), \
                patch("release.ensure_tag") as promote, self.assertRaises(ValueError):
            release.promote_images(REPOSITORY, "v0.0.9", SOURCE)
        promote.assert_not_called()

    def test_promotion_diagnostic_identifies_partial_version_component(self):
        with patch("release.select_release", return_value="v0.0.13"), \
                patch("release.inspect_image", side_effect=[DIGEST, None, DIGEST, None, DIGEST, OTHER_DIGEST]), \
                patch("release.ensure_tag") as promote, self.assertRaises(release.ReleaseError) as caught:
            release.promote_images(REPOSITORY, "v0.0.13", SOURCE)
        self.assertIn("Promotion preflight for importer at v0.0.13", str(caught.exception))
        self.assertIn("digests differ", str(caught.exception))
        promote.assert_not_called()

    def test_revision_diagnostic_reports_only_valid_commit_ids(self):
        for revision, expected in (("d" * 40, "d" * 40), ("synthetic-secret-value", "missing or invalid")):
            config = {"architecture": "amd64", "os": "linux", "config": {
                "Labels": {"org.opencontainers.image.revision": revision},
            }}
            with self.subTest(revision=revision), patch("release.run", side_effect=[
                    self.result(json.dumps({"digest": DIGEST})), self.result(json.dumps(config)),
            ]), self.assertRaises(release.ReleaseError) as caught:
                release.inspect_image(f"{IMAGE}:v0.0.13", SOURCE)
            self.assertIn(f"expected {SOURCE}, found {expected}", str(caught.exception))
            self.assertNotIn("synthetic-secret-value", str(caught.exception))

    def test_occupancy_check_accepts_other_valid_revision_but_not_missing_labels(self):
        for revision in ("d" * 40, None):
            config = {"architecture": "amd64", "os": "linux", "config": {
                "Labels": {"org.opencontainers.image.revision": revision},
            }}
            with self.subTest(revision=revision), patch("release.run", side_effect=[
                    self.result(json.dumps({"digest": OTHER_DIGEST})), self.result(json.dumps(config)),
            ]):
                if revision is None:
                    with self.assertRaises(release.ReleaseError):
                        release.inspect_image(f"{IMAGE}:v0.0.13", None)
                else:
                    self.assertEqual(release.inspect_image(f"{IMAGE}:v0.0.13", None), OTHER_DIGEST)

    def test_missing_source_prevents_all_version_promotions(self):
        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.inspect_image", side_effect=[DIGEST, None, DIGEST, None, None]), \
                patch("release.ensure_tag") as promote, self.assertRaises(ValueError):
            release.promote_images(REPOSITORY, "v0.0.9", SOURCE)
        promote.assert_not_called()

    def test_promotion_uses_original_digest_and_verifies_result(self):
        with patch("release.inspect_image", side_effect=[None, DIGEST]), patch("release.run") as commands:
            release.ensure_tag(IMAGE, "v0.0.9", DIGEST, SOURCE)
        commands.assert_called_once_with([
            "docker", "buildx", "imagetools", "create", "--prefer-index=false",
            "--tag", f"{IMAGE}:v0.0.9", f"{IMAGE}@{DIGEST}",
        ])
        with patch("release.inspect_image", side_effect=[None, OTHER_DIGEST]), \
                patch("release.run"), self.assertRaises(ValueError):
            release.ensure_tag(IMAGE, "v0.0.9", DIGEST, SOURCE)

    def test_new_migrator_build_uses_its_dockerfile_and_source_label(self):
        def build(command):
            metadata = Path(command[command.index("--metadata-file") + 1])
            metadata.write_text(json.dumps({"containerimage.digest": DIGEST}))

        with patch("release.inspect_image", return_value=DIGEST), \
                patch("release.run", side_effect=build) as commands:
            self.assertEqual(release.build_candidate(REPOSITORY, "migrator", SOURCE), DIGEST)
        build = commands.call_args_list[0].args[0]
        self.assertEqual(build[:3], ["docker", "buildx", "build"])
        self.assertIn("texinroistot-server/Dockerfile.migrator", build)
        self.assertIn(f"org.opencontainers.image.revision={SOURCE}", build)
        self.assertIn(f"type=image,name=ghcr.io/{REPOSITORY}-migrator,push-by-digest=true,name-canonical=true,push=true", build)
        self.assertNotIn("--tag", build)

    def test_failed_build_never_promotes_a_version(self):
        with patch("release.inspect_image", return_value=None), \
                patch("release.run", side_effect=ValueError("build failed")) as commands, \
                self.assertRaises(ValueError):
            release.publish_image(REPOSITORY, "backend", SOURCE)
        self.assertEqual(commands.call_count, 1)
        self.assertEqual(commands.call_args.args[0][2], "build")

    def test_tag_verification_refuses_different_digest(self):
        with patch("release.inspect_image", return_value=OTHER_DIGEST), \
                patch("release.run") as commands, self.assertRaises(ValueError):
            release.ensure_tag(IMAGE, "v0.0.9", DIGEST, SOURCE)
        commands.assert_not_called()


class ReleaseTests(unittest.TestCase):
    def setUp(self):
        reservation = patch("release.require_reservation")
        self.reservation = reservation.start()
        self.addCleanup(reservation.stop)

    def test_missing_image_prevents_any_github_write(self):
        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.release_tags", return_value=["v0.0.8"]), \
                patch("release.inspect_image", side_effect=[DIGEST] * 6 + [None]), \
                patch("release.github_api") as api, self.assertRaises(ValueError):
            release.publish_release(REPOSITORY, "v0.0.9", SOURCE)
        api.assert_not_called()

    def test_release_created_only_after_all_four_images_verified(self):
        calls = []

        def inspect(reference, source):
            calls.append(reference)
            return DIGEST

        def api(repository, path, payload=None, **kwargs):
            self.assertEqual(len(calls), 8)
            if path == "releases/tags/v0.0.8":
                return {"tag_name": "v0.0.8", "draft": False, "prerelease": False}
            if path == "releases/generate-notes":
                self.assertEqual(payload["previous_tag_name"], "v0.0.8")
                return {"body": "Synthetic release notes"}
            return None

        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.release_tags", return_value=["v0.0.8", "v0.0.9"]), \
                patch("release.inspect_image", side_effect=inspect), \
                patch("release.github_api", side_effect=api) as operations:
            self.assertTrue(release.publish_release(REPOSITORY, "v0.0.9", SOURCE))
        self.assertFalse(any(call.args[1] == "git/refs" for call in operations.call_args_list))
        payload = operations.call_args.args[2]
        for component in release.COMPONENTS:
            self.assertIn(f"ghcr.io/{REPOSITORY}-{component}@{DIGEST}", payload["body"])
        self.assertFalse(payload["draft"])
        self.assertEqual(payload["target_commitish"], SOURCE)

    def test_completed_release_retry_performs_no_github_writes(self):
        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.release_tags", return_value=["v0.0.8", "v0.0.9"]), \
                patch("release.inspect_image", return_value=DIGEST), \
                patch("release.github_api", return_value={"tag_name": "v0.0.9", "draft": False, "prerelease": False}) as api:
            self.assertTrue(release.publish_release(REPOSITORY, "v0.0.9", SOURCE))
        api.assert_called_once_with(REPOSITORY, "releases/tags/v0.0.9", missing_ok=True)

    def test_notes_skip_reserved_tags_without_completed_releases(self):
        def api(repository, path, payload=None, **kwargs):
            if path == "releases/tags/v0.0.12":
                return {"tag_name": "v0.0.12", "draft": False, "prerelease": False}
            if path == "releases/generate-notes":
                self.assertEqual(payload["previous_tag_name"], "v0.0.12")
                return {"body": "Changes since the last completed release"}
            return None

        with patch("release.release_tags", return_value=["v0.0.12", "v0.0.13", "v0.0.14"]), \
                patch("release.inspect_image", return_value=DIGEST), \
                patch("release.github_api", side_effect=api) as operations:
            self.assertTrue(release.publish_release(REPOSITORY, "v0.0.14", SOURCE))
        operations.assert_any_call(REPOSITORY, "releases/tags/v0.0.13", missing_ok=True)
        operations.assert_any_call(REPOSITORY, "releases/tags/v0.0.12", missing_ok=True)

    def test_old_release_retry_is_not_handoff_eligible(self):
        with patch("release.select_release", return_value="v0.0.8"), \
                patch("release.release_tags", return_value=["v0.0.8", "v0.0.9"]), \
                patch("release.inspect_image", return_value=DIGEST), \
                patch("release.github_api", return_value={"tag_name": "v0.0.8", "draft": False, "prerelease": False}):
            self.assertFalse(release.publish_release(REPOSITORY, "v0.0.8", SOURCE))

    def test_changed_allocation_fails_before_registry_or_github_operations(self):
        self.reservation.side_effect = release.ReleaseError("Reservation belongs to another commit")
        with patch("release.select_release", return_value="v0.0.10"), \
                patch("release.inspect_image") as inspect, patch("release.github_api") as api, \
                self.assertRaises(ValueError):
            release.publish_release(REPOSITORY, "v0.0.9", SOURCE)
        inspect.assert_not_called()
        api.assert_not_called()

    def test_api_only_treats_404_as_missing(self):
        for status in (401, 403, 404, 429, 500):
            failure = error.HTTPError("https://api.github.com/synthetic", status, "synthetic", {}, None)
            with self.subTest(status=status), patch.dict(os.environ, {"GH_TOKEN": "synthetic"}, clear=True), \
                    patch("release.request.build_opener") as opener:
                opener.return_value.open.side_effect = failure
                if status == 404:
                    self.assertIsNone(release.github_api(REPOSITORY, "releases/tags/v0.0.9", missing_ok=True))
                else:
                    with self.assertRaises(ValueError):
                        release.github_api(REPOSITORY, "releases/tags/v0.0.9", missing_ok=True)

    def test_cli_failure_does_not_print_command_diagnostics_or_credentials(self):
        output = io.StringIO()
        with patch("sys.argv", ["release", "plan", "--repository", REPOSITORY, "--source-sha", SOURCE]), \
                patch("release.run", side_effect=OSError("synthetic-sensitive-diagnostic")), \
                contextlib.redirect_stderr(output):
            self.assertEqual(release.main(), 1)
        self.assertNotIn("synthetic-sensitive-diagnostic", output.getvalue())

    def test_github_diagnostic_reaches_cli_without_raw_body_or_token(self):
        token = "synthetic-secret-token"
        body = io.BytesIO(json.dumps({
            "message": "Validation Failed",
            "errors": [{"code": "already_exists", "message": token, "value": "private-value"}],
            "documentation_url": "https://private.example.invalid/" + token,
        }).encode())
        failure = error.HTTPError("https://api.github.com/synthetic", 422, token, {}, body)
        output = io.StringIO()
        with patch("sys.argv", ["release", "publish", "--repository", REPOSITORY,
                                "--source-sha", SOURCE, "--release-tag", "v0.0.9"]), \
                patch.dict(os.environ, {"GH_TOKEN": token}, clear=True), \
                patch("release.run", return_value=subprocess.CompletedProcess([], 0, SOURCE, "")), \
                patch("release.publish_release", side_effect=lambda *args: release.github_api(REPOSITORY, "git/refs", {})), \
                patch("release.request.build_opener") as opener, contextlib.redirect_stderr(output):
            opener.return_value.open.side_effect = failure
            self.assertEqual(release.main(), 1)
        self.assertIn("Release publish failed", output.getvalue())
        self.assertIn("create Git tag (HTTP 422): Validation Failed", output.getvalue())
        self.assertIn("already_exists", output.getvalue())
        for sensitive in (token, "private-value", "private.example.invalid"):
            self.assertNotIn(sensitive, output.getvalue())
        self.assertTrue(body.closed)

    def test_github_error_operation_and_status_are_reported(self):
        for path, operation in (("git/refs", "create Git tag"),
                                ("releases/generate-notes", "generate release notes"),
                                ("releases", "create GitHub Release"),
                                ("releases/tags/v0.0.9", "look up GitHub Release")):
            body = io.BytesIO(b'{"message":"Resource not accessible by integration"}')
            failure = error.HTTPError("https://api.github.com/synthetic", 403, "synthetic", {}, body)
            with self.subTest(path=path), patch.dict(os.environ, {"GH_TOKEN": "synthetic"}, clear=True), \
                    patch("release.request.build_opener") as opener, self.assertRaises(release.ReleaseError) as caught:
                opener.return_value.open.side_effect = failure
                release.github_api(REPOSITORY, path)
            self.assertIn(operation, str(caught.exception))
            self.assertIn("HTTP 403", str(caught.exception))
            self.assertIn("Resource not accessible by integration", str(caught.exception))
            self.assertTrue(body.closed)

    def test_github_error_body_is_bounded_and_unknown_content_is_omitted(self):
        bodies = [b'<html>synthetic-secret</html>', b'[]', b'{"message":["synthetic-secret"]}',
                  b'{"message":"synthetic-secret\\n::warning::injected"}',
                  b'{"message":"Validation Failed","errors":[{"code":{}},"synthetic-secret"]}',
                  b' ' * 8192 + b'{"message":"synthetic-secret"}']
        for payload in bodies:
            body = io.BytesIO(payload)
            failure = error.HTTPError("https://api.github.com/synthetic", 500, "synthetic-secret", {}, body)
            with self.subTest(payload=payload[:60]), patch.dict(os.environ, {"GH_TOKEN": "synthetic-secret"}, clear=True), \
                    patch("release.request.build_opener") as opener, \
                    patch.object(failure, "read", wraps=failure.read) as read, \
                    self.assertRaises(release.ReleaseError) as caught:
                opener.return_value.open.side_effect = failure
                release.github_api(REPOSITORY, "releases", {})
            read.assert_called_once_with(8192)
            self.assertIn("HTTP 500", str(caught.exception))
            self.assertNotIn("synthetic-secret", str(caught.exception))
            self.assertNotIn("::warning::", str(caught.exception))
            self.assertTrue(body.closed)

    def test_github_body_read_failure_preserves_status_and_closes_response(self):
        body = io.BytesIO(b"synthetic-secret")
        failure = error.HTTPError("https://api.github.com/synthetic", 502, "synthetic-secret", {}, body)
        with patch.dict(os.environ, {"GH_TOKEN": "synthetic"}, clear=True), \
                patch("release.request.build_opener") as opener, \
                patch.object(failure, "read", side_effect=OSError("synthetic-secret")), \
                self.assertRaises(release.ReleaseError) as caught:
            opener.return_value.open.side_effect = failure
            release.github_api(REPOSITORY, "releases", {})
        self.assertIn("HTTP 502", str(caught.exception))
        self.assertNotIn("synthetic-secret", str(caught.exception))
        self.assertTrue(body.closed)

    def test_expected_release_guard_is_visible_but_arbitrary_value_errors_are_not(self):
        for failure, visible in ((release.ReleaseError("Immutable image tag already has a different digest"), True),
                                 (ValueError("synthetic-sensitive-diagnostic"), False)):
            output = io.StringIO()
            with self.subTest(visible=visible), \
                    patch("sys.argv", ["release", "plan", "--repository", REPOSITORY, "--source-sha", SOURCE]), \
                    patch("release.run", side_effect=failure), contextlib.redirect_stderr(output):
                self.assertEqual(release.main(), 1)
            self.assertEqual(str(failure) in output.getvalue(), visible)

    def test_cli_rejects_checkout_other_than_tested_commit(self):
        with patch("sys.argv", ["release", "publish", "--repository", REPOSITORY,
                                "--source-sha", SOURCE, "--release-tag", "v0.0.9"]), \
                patch("release.run", return_value=subprocess.CompletedProcess([], 0, "d" * 40, "")), \
                patch("release.publish_release") as publish, contextlib.redirect_stderr(io.StringIO()):
            self.assertEqual(release.main(), 1)
        publish.assert_not_called()


if __name__ == "__main__":
    unittest.main()
