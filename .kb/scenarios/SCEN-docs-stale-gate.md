---
title: Given a commit changes internal/ledger/ledger.go, which docs/concepts/ledger.md covers, and no page covering i
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:32:41.656Z'
id: SCEN-docs-stale-gate
type: scenario
---
Given a commit changes internal/ledger/ledger.go, which docs/concepts/ledger.md covers, and no page covering it changes in the push, when the staleness gate runs, then it fails naming the file and the pages to update; with a Docs-Unchanged trailer, or with the page changed later in the push, it passes. Test files and testdata never trigger it.
