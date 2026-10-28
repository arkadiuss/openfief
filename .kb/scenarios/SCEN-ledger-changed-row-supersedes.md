---
title: A changed row becomes a new version and the old one is kept
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:10.844Z'
id: SCEN-ledger-changed-row-supersedes
type: scenario
---
Given versions 1 and 2 of transaction b-1 in the log, when the row arrives with content that differs from version 2, then the decision is a new version 3 that supersedes version 2, and versions 1 and 2 stay in the log; when it arrives equal to version 2, then nothing changes.