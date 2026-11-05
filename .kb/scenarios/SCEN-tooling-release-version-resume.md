---
title: A re-run of publish accepts the tag it already created at the release commit
status: active
origin:
  kind: agent
  recorded_at: '2026-11-05T21:53:58.173Z'
id: SCEN-tooling-release-version-resume
type: scenario
---
Given tag 2026.10.0-alpha.1 at the release commit and a later commit, when task release:check runs for it with COMMIT set to the release commit, then it passes and says it resumes; with COMMIT at the later commit, an unknown commit, or no COMMIT at all, it fails because the tag already exists.
