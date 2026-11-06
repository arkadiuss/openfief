---
title: openfief version names the release and commit, or dev
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T16:08:50.336Z'
id: SCEN-tooling-cli-version
type: scenario
---
Given a CLI the release built for 2026.11.0-alpha.2 from a clean checkout of commit C, when the owner runs openfief version, then it prints openfief 2026.11.0-alpha.2 commit C. A local build prints dev instead of a version, a build from a modified tree marks the commit -modified, a build without version control information prints commit unknown, and any argument is a usage error.
