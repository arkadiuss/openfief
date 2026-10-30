---
title: A new version of the pack's rules revalues every account once
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:01.555Z'
id: SCEN-valuation-rules-version-revalues
type: scenario
---
Given a recorded rules version, when the core starts with the same version, then nothing is enqueued; when the version differs or none was recorded, then every account is enqueued from its first transaction in the same transaction that records it.