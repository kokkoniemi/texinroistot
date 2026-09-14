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


class ImageTests(unittest.TestCase):
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
    def test_missing_image_prevents_any_github_write(self):
        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.release_tags", return_value=["v0.0.8"]), \
                patch("release.inspect_image", side_effect=[DIGEST] * 6 + [None]), \
                patch("release.github_api") as api, self.assertRaises(ValueError):
            release.publish_release(REPOSITORY, "v0.0.9", SOURCE)
        api.assert_not_called()

    def test_tag_and_release_created_only_after_all_four_images_verified(self):
        calls = []

        def inspect(reference, source):
            calls.append(reference)
            return DIGEST

        def api(repository, path, payload=None, **kwargs):
            self.assertEqual(len(calls), 8)
            if path == "releases/generate-notes":
                self.assertEqual(payload["previous_tag_name"], "v0.0.8")
                return {"body": "Synthetic release notes"}
            return None

        with patch("release.select_release", return_value="v0.0.9"), \
                patch("release.release_tags", return_value=["v0.0.8"]), \
                patch("release.inspect_image", side_effect=inspect), \
                patch("release.github_api", side_effect=api) as operations:
            self.assertTrue(release.publish_release(REPOSITORY, "v0.0.9", SOURCE))
        self.assertEqual(operations.call_args_list[0].args[1:], (
            "git/refs", {"ref": "refs/tags/v0.0.9", "sha": SOURCE},
        ))
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

    def test_old_release_retry_is_not_handoff_eligible(self):
        with patch("release.select_release", return_value="v0.0.8"), \
                patch("release.release_tags", return_value=["v0.0.8", "v0.0.9"]), \
                patch("release.inspect_image", return_value=DIGEST), \
                patch("release.github_api", return_value={"tag_name": "v0.0.8", "draft": False, "prerelease": False}):
            self.assertFalse(release.publish_release(REPOSITORY, "v0.0.8", SOURCE))

    def test_changed_allocation_fails_before_registry_or_github_operations(self):
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

    def test_cli_rejects_checkout_other_than_tested_commit(self):
        with patch("sys.argv", ["release", "publish", "--repository", REPOSITORY,
                                "--source-sha", SOURCE, "--release-tag", "v0.0.9"]), \
                patch("release.run", return_value=subprocess.CompletedProcess([], 0, "d" * 40, "")), \
                patch("release.publish_release") as publish, contextlib.redirect_stderr(io.StringIO()):
            self.assertEqual(release.main(), 1)
        publish.assert_not_called()


if __name__ == "__main__":
    unittest.main()
