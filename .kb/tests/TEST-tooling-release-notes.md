---
title: release-notes.sh groups commits since the right previous release
status: active
sourceFile: scripts/tests/release-notes.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:58.837Z'
id: TEST-tooling-release-notes
type: test
verification_scope: integration
---
Builds a throwaway repository with release tags and Conventional Commits and compares the notes for a stable release, a beta and a first release with the expected Markdown.
