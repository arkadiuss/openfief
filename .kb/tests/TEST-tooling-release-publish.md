---
title: release-publish.sh publishes once and a re-run finishes only what is missing
status: active
sourceFile: scripts/tests/release-publish.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-05T21:54:47.163Z'
id: TEST-tooling-release-publish
type: test
---
Runs scripts/release-publish.sh in throwaway repositories with a stand-in gh that logs every write and runs --jq through gojq: a first publish tags HEAD, creates the release with every asset (a pre-release unless stable) and opens the pull request; a re-run keeps a tag at HEAD and a complete release and opens only the pull request; uploads only the missing assets and replaces an unfinished one; opens no pull request when an open or merged one exists but does when only a closed one does; writes nothing when all is done. A tag at another commit, a missing bundle file or CLI, and a release it cannot read fail.
