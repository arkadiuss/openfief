---
title: Names are kept per ISIN, append-only; only the owner renames
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:04.234Z'
id: SCEN-ledger-instrument-names
type: scenario
---
Given deliveries naming an ISIN twice, when the ledger stores them, then the first stays; the owner's rename is appended and wins; a read session, a malformed ISIN or an empty name is refused.