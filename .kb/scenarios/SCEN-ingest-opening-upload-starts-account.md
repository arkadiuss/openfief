---
title: An opening upload starts its account at once
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T16:48:20.445Z'
id: SCEN-ingest-opening-upload-starts-account
type: scenario
---
Given an account not started, when the owner uploads an opening statement whose positions all wait for their tickers and no cash, then the account's go-live is recorded with the upload, the next upload is a transactions batch that books nothing dated on or before it, and each confirmed position joins the opening on that date.