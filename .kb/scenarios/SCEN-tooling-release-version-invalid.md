---
title: An invalid release version is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:47.942Z'
id: SCEN-tooling-release-version-invalid
type: scenario
---
Given channel stable and version v2026.11.0, 2026.01.0 or 2026.11.0-beta.1, or channel beta and version 2026.11.0 or 2026.11.0-beta.0, or a version whose tag already exists, when task release:check runs, then it exits non-zero. 2026.11.0 for stable and 2026.11.0-beta.1 for beta pass.
