---
title: Given a directory with non-test Go files that no page covers, when the staleness gate runs
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:32:42.910Z'
id: SCEN-docs-packages-covered
type: scenario
---
Given a directory with non-test Go files that no page covers, when the staleness gate runs, then it fails naming the directory.

Without this check a new package could arrive with no page at all, and the per-change gate would never notice, because it only looks at files some page already covers. The author verifies shipped code by reading the docs, so undocumented code is invisible to that review.
