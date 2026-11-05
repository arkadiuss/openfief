---
title: Release notes carry the version's known limits from its file
status: active
origin:
  kind: agent
  recorded_at: '2026-11-05T21:55:34.659Z'
id: SCEN-tooling-release-notes-known-limits
type: scenario
---
Given .github/release-notes/2026.11.0-alpha.2/known-limits.md committed at the ref, when task release:notes runs for 2026.11.0-alpha.2, then a Known limits section with the file's Markdown follows the intro line and comes before the changes; a version without its own file gets no such section, and an empty file makes the notes fail.
