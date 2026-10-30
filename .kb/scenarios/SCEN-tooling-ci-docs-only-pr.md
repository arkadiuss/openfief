---
title: A docs-only pull request skips the e2e and PostgreSQL jobs
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:08.502Z'
id: SCEN-tooling-ci-docs-only-pr
type: scenario
---
Given a pull request that changes only documentation or requirements, when CI plans the run, then the end-to-end and PostgreSQL jobs are skipped and task check still runs in full; given a push to main, both jobs always run.