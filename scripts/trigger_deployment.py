#!/usr/bin/env python3

import argparse
import json
import os
import re
import sys
from urllib import error, parse, request


class NoRedirect(request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        if fp is not None:
            fp.close()
        raise error.URLError("Redirect refused")


def trigger_deployment(endpoint, token, release_tag, source_sha):
    if not re.fullmatch(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)", release_tag):
        raise ValueError("A strict release tag is required")
    if not re.fullmatch(r"[0-9a-f]{40}", source_sha):
        raise ValueError("A full source commit SHA is required")
    parsed = parse.urlsplit(endpoint)
    if (parsed.scheme != "https" or not parsed.hostname or parsed.username or parsed.password
            or parsed.query or parsed.fragment
            or not re.fullmatch(r"/api/v4/projects/[0-9]+/trigger/pipeline", parsed.path)):
        raise ValueError("Configure an HTTPS GitLab project pipeline-trigger endpoint")
    if not token:
        raise ValueError("Configure the dedicated pipeline trigger credential")
    payload = parse.urlencode({
        "token": token,
        "ref": "main",
        "inputs[release_tag]": release_tag,
        "inputs[source_sha]": source_sha,
    }).encode()
    operation = request.Request(endpoint, data=payload, method="POST",
                                headers={"Content-Type": "application/x-www-form-urlencoded"})
    opener = request.build_opener(NoRedirect())
    with opener.open(operation, timeout=30) as response:
        if response.status != 201:
            raise ValueError("Private pipeline did not acknowledge creation")
        result = json.load(response)
        if not isinstance(result, dict) or type(result.get("id")) is not int or result["id"] <= 0:
            raise ValueError("Private pipeline returned an invalid acknowledgement")


def main():
    parser = argparse.ArgumentParser(description="Request a private deployment after release publication")
    parser.add_argument("--release-tag", required=True)
    parser.add_argument("--source-sha", required=True)
    args = parser.parse_args()
    try:
        trigger_deployment(os.environ.get("GITLAB_TRIGGER_URL", ""),
                           os.environ.get("GITLAB_TRIGGER_TOKEN", ""), args.release_tag, args.source_sha)
    except (ValueError, OSError, error.URLError):
        print("Private deployment request failed; check configuration and private pipeline status.", file=sys.stderr)
        return 1
    print("Private deployment pipeline accepted the release request; rollout status is in GitLab.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
