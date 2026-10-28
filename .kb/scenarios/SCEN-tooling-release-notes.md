---
title: Release notes group the commits since the previous release
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:59.873Z'
id: SCEN-tooling-release-notes
type: scenario
---
Given tags 2026.10.0 and 2026.11.0-beta.1 and later commits of types feat, fix, perf, a breaking change and a non-conventional subject, when task release:notes runs for 2026.11.0, then it lists everything since 2026.10.0 under Breaking changes, Features, Fixes, Performance and Other changes; for 2026.11.0-beta.2 it starts after 2026.11.0-beta.1; with no earlier release it covers the whole history.
