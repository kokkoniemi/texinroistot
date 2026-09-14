import contextlib
import io
import json
import unittest
from unittest.mock import patch
from urllib import error, parse

from trigger_deployment import NoRedirect, main, trigger_deployment


ENDPOINT = "https://gitlab.example.invalid/api/v4/projects/123/trigger/pipeline"
TOKEN = "synthetic-trigger-value"
SOURCE_SHA = "a" * 40


class TriggerDeploymentTests(unittest.TestCase):
    def test_sends_only_validated_release_inputs_and_trigger_credential(self):
        response = io.StringIO(json.dumps({"id": 1}))
        response.status = 201
        with patch("trigger_deployment.request.build_opener") as factory:
            factory.return_value.open.return_value = response
            trigger_deployment(ENDPOINT, TOKEN, "v0.0.9", SOURCE_SHA)
        operation = factory.return_value.open.call_args.args[0]
        self.assertEqual(operation.full_url, ENDPOINT)
        self.assertEqual(operation.method, "POST")
        self.assertEqual(parse.parse_qs(operation.data.decode()), {
            "token": [TOKEN], "ref": ["main"],
            "inputs[release_tag]": ["v0.0.9"], "inputs[source_sha]": [SOURCE_SHA],
        })
        self.assertIsInstance(factory.call_args.args[0], NoRedirect)

    def test_rejects_unsafe_endpoints_and_invalid_releases_without_network(self):
        invalid = [
            (ENDPOINT.replace("https:", "http:"), "v0.0.9", SOURCE_SHA),
            (ENDPOINT + "?token=value", "v0.0.9", SOURCE_SHA),
            (ENDPOINT.replace("https://", "https://user:pass@"), "v0.0.9", SOURCE_SHA),
            (ENDPOINT, "latest", SOURCE_SHA),
            (ENDPOINT, "v01.2.3", SOURCE_SHA),
            (ENDPOINT, "v1.2.3-rc1", SOURCE_SHA),
            (ENDPOINT, "v0.0.9", "short"),
        ]
        with patch("trigger_deployment.request.build_opener") as factory:
            for endpoint, tag, source_sha in invalid:
                with self.subTest(endpoint=endpoint, tag=tag), self.assertRaises(ValueError):
                    trigger_deployment(endpoint, TOKEN, tag, source_sha)
            factory.assert_not_called()

    def test_refuses_redirect_instead_of_forwarding_token(self):
        from urllib.request import Request
        with self.assertRaises(error.URLError):
            NoRedirect().redirect_request(Request(ENDPOINT), None, 307, "redirect", {}, "https://other.invalid")

    def test_failure_output_does_not_expose_endpoint_token_or_response(self):
        output = io.StringIO()
        with patch("sys.argv", ["trigger", "--release-tag", "v0.0.9", "--source-sha", SOURCE_SHA]), \
                patch.dict("os.environ", {"GITLAB_TRIGGER_URL": ENDPOINT, "GITLAB_TRIGGER_TOKEN": TOKEN}, clear=True), \
                patch("trigger_deployment.request.build_opener", side_effect=error.URLError(ENDPOINT + TOKEN)), \
                contextlib.redirect_stderr(output):
            self.assertEqual(main(), 1)
        self.assertNotIn(TOKEN, output.getvalue())
        self.assertNotIn(ENDPOINT, output.getvalue())

    def test_rejects_unacknowledged_pipeline(self):
        response = io.StringIO("{}")
        response.status = 201
        with patch("trigger_deployment.request.build_opener") as factory:
            factory.return_value.open.return_value = response
            with self.assertRaises(ValueError):
                trigger_deployment(ENDPOINT, TOKEN, "v0.0.9", SOURCE_SHA)


if __name__ == "__main__":
    unittest.main()
