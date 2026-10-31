---
title: A changed decision supersedes and the same decision appends nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T06:05:41.166Z'
id: SCEN-ledger-transfer-correction
type: scenario
---
Given a confirmed pair stored as version 1, when the owner confirms it again nothing is appended; when the owner rejects it, version 2 is appended, becomes the current link and version 1 stays stored; the database refuses changing, deleting or truncating stored links.