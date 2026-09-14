# Release and deployment boundary

This public repository owns application source, migrations, tests, and image builds. The private infrastructure repository owns deployment bundles, host configuration, operational plans, deployment jobs, and their logs. Do not add live deployment files or host details here. Git ignore rules prevent accidental additions, but do not make tracked files private or erase published history.

## Automatic release flow

`ci.yml` is the only push entry point. Pull requests run checks only; pushes to `main` and manual CI runs on `main` run this sequence:

1. A push to `main` passes application and migration checks.
2. The reusable `images.yml` workflow builds backend, frontend, importer, and migrator images and publishes their immutable `sha-<full SHA>` tags.
3. After verifying all four source images, it reserves the next available strict SemVer version by creating its Git tag for the tested commit. It reuses an existing reservation for the same commit and skips conflicting image-only versions left by older workflows.
4. It promotes all four images to the reserved `vX.Y.Z` tag, verifies revision labels and matching digests, then creates the GitHub Release.
5. If explicitly enabled, a final GitHub job invokes `scripts/trigger_deployment.py` with the release tag and source commit.
6. GitLab runs the private rollout pipeline, validating the requested release before deployment.

The release path is implemented; no remote publication or deployment is performed by editing these files. Merging this workflow to `main` enables release publication. Deployment handoff remains off unless the operator sets the GitHub repository variable `ENABLE_GITLAB_HANDOFF=true`. The private pipeline has its own separate enablement gate; keep both off until rehearsal and the database transition succeed.

Whole CI runs share a branch-specific concurrency group with `cancel-in-progress: false` and `queue: max`. GitHub permits up to 100 pending runs; overflow is canceled. Queue order follows when runs start waiting, not necessarily commit order. Version selection rejects untagged commits older than the latest release, and old release retries cannot trigger automatic rollback. See [GitHub concurrency semantics](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency).

There is no version commit, independent tag-triggered build, or mutable `latest`, `main`, major, or minor tag publication. Use full release tags. Existing convenience tags from the previous workflow are no longer updated.

## Image and retry safety

`scripts/release.py` uses numeric SemVer ordering, beginning at `v0.0.1` when there are no strict release tags. Annotated and lightweight tags are supported. More than one release tag on the same commit requires operator review. Tags must be protected against manual moves/deletion; image write access must be limited to trusted publishers. The workflow refuses overwrites, but does not make GHCR tags immutable against other writers.

A Git tag is now a durable version reservation, not proof that publication finished. A completed GitHub Release is the publication marker. If promotion fails, the reservation remains so a later queued commit cannot reuse its version. Reservation checks up to 100 candidate versions and fails closed on registry errors or a conflicting existing Git reservation. No image tag is overwritten or deleted to make a version available.

Each build first pushes an untagged, content-addressed candidate. This also creates a new image package before checking named tags; authentication failures are never interpreted as proof that an image is absent. Only a confirmed missing tag can be created. Both named tags must resolve to the same digest and carry the tested commit's revision label. Builds target `linux/amd64`; provenance attestations are deferred to the hardening phase. See [Docker's image exporter](https://docs.docker.com/build/exporters/image-registry/).

Retries can rebuild candidates, but retain existing source/release digests even if a base image has changed. Image jobs publish only source-SHA tags. Promotion requires an existing Git reservation for the exact commit and preflights all four source and version tags before assigning any version tag, so a partial earlier release cannot be mixed with a later commit. A partial promotion is completed from its already-published source digests. All four components are verified again before GitHub Release creation. Generated release notes include the source SHA and four image digests, and start from the previous completed release rather than an incomplete reservation.

Recovery rules:

- Failed checks publish nothing. Failed image jobs may leave partial image tags or untagged candidates, but create no GitHub Release and perform no handoff.
- Retry the failed run for the same SHA to finish a partial release using its reservation. Later commits allocate newer versions even when an earlier reservation has no completed release. Do not delete or overwrite tags.
- Legacy image-only versions are reused only when all existing component digests match this commit's source images. Otherwise they are skipped and preserved; the workflow does not create a Git tag for an orphan belonging to another commit.
- If only GitHub Release creation fails after the Git tag is created, retry reuses the tag and completes the release.
- Rerunning an already-published release does not allocate another version. A superseded release cannot trigger handoff. Application rollback remains an explicit operation in the private repository, keeping migrations forward-only.
- If handoff times out, check private GitLab status before retrying: the request may have been accepted. A repeated request can create another pipeline; the trigger API does not provide exactly-once delivery.

### Recovering from the image-only version collision

For example, if `v0.0.13` images exist but no Git tag or GitHub Release exists:

1. Commit and merge the reservation fix and diagnostic changes. Do not force-create `v0.0.13` on the newest commit, remove images, or move tags.
2. Let the new `main` run execute, or start the **CI** workflow manually on `main`. Re-running an old failed run still uses its old workflow/scripts and does not apply this fix. Any in-progress old run must finish or be canceled before the corrected run proceeds; normal concurrency serializes the runs.
3. The reservation job checks the existing images. If they belong to a different build, it logs that `v0.0.13` is skipped and reserves the next free version, normally `v0.0.14`. Other occupied versions are skipped in the same way.
4. Confirm the new GitHub Release and all four matching image tags. Leave the orphan images untouched for later, explicitly authorized cleanup. A missing GitHub Release still means incomplete publication, even when a Git reservation exists.

Keep handoff disabled during recovery unless you intentionally want the recovered release deployed. No credential changes are required for the collision fix; the reservation job uses the workflow token's existing repository-write permission, moved earlier in the pipeline.

## Operator setup

The operator configures `GITLAB_TRIGGER_URL` and a dedicated `GITLAB_TRIGGER_TOKEN` as GitHub Actions secrets. The URL must be an HTTPS GitLab pipeline-trigger endpoint for the private project. The helper sends validated `release_tag` and `source_sha` pipeline inputs to the private repository's `main` branch. It rejects redirects and never prints the token, endpoint, or response body. Trigger acceptance is not deployment success; the final outcome is recorded in private GitLab jobs.

Only the trusted post-release `main` job may use the trigger credential. Do not invoke the helper in pull-request jobs. SSH/VPN credentials and host settings stay outside this repository and its Actions jobs. Operator configuration does not require an agent to inspect existing credentials.

Quality jobs have read-only repository permissions. Image and promotion jobs add package writes. Version reservation and release creation have repository writes and read-only package access. The handoff job has repository read access only. The reusable-workflow caller grants the maximum needed permissions, which its individual jobs reduce. Checkout does not persist credentials.

Ensure repository rules allow the scoped workflow token to create release tags and releases, while preventing other unreviewed tag changes. New GHCR packages may initially be private: the operator must review visibility and package access for all four components, including the new migrator. Host-side access is configured privately, not by this workflow.

Third-party actions are pinned to reviewed commits. `.github/dependabot.yml` proposes weekly action updates for review; do not auto-merge privileged workflow changes.

The receiving GitLab pipeline must validate inputs, restrict execution to its protected deployment branch, serialize rollouts, and verify that image source labels match the requested source commit. Its rollout logic must not invoke infrastructure apply commands. Keep deployment disabled until a disposable-host rehearsal and the database transition have succeeded.

## Local validation

### Release failure diagnostics

Expected release-guard failures include their reason in the CLI error. GitHub API failures include the operation (tag creation, release lookup, notes generation, or release creation), HTTP status, recognized GitHub error messages, and recognized validation codes. Raw response bodies, arbitrary server messages, response URLs, and credentials are not printed. Unrecognized, malformed, or oversized bodies are omitted; unexpected exceptions retain generic output.

Use the failed job/step and this diagnostic to investigate before changing action pins or token permissions. A merged Dependabot update alone does not establish the cause. Do not paste credentials or private deployment logs into public issues. A diagnostic change is not a repair of the failed release itself; the original retry and partial-promotion rules still apply.

Validate release allocation, immutable-tag guards, partial failures, stale retries, and the handoff helper without network access or real credentials:

```bash
python3 -B -m unittest discover -s scripts -p 'test_*.py'
```

These tests use temporary Git repositories and mocked registry/GitHub/GitLab operations. They do not prove live registry permissions or end-to-end CI execution. The first observed release must confirm all four tags/digests and notes while handoff remains disabled.

References: [GitLab pipeline triggers and validated inputs](https://docs.gitlab.com/ci/triggers/), [GitHub reusable workflows](https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows).
