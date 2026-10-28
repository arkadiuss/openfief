---
title: Re-importing a broker export appends nothing, a corrected row appends one version
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:22.223Z'
id: SCEN-ledger-app-reimport
type: scenario
---
Given an account whose deposit and buy were imported once, when the same rows are imported again, even with postings in another order or amounts written with trailing zeros, then nothing is appended and the import reports both unchanged; when the buy's row comes back with 12 units instead of 10, then version 2 is appended and version 1 stays stored, unchanged. Assumes the institution's ID for a row is stable across exports.