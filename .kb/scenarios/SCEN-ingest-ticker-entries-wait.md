---
title: Ticker entries wait for their ISIN
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:09.630Z'
id: SCEN-ingest-ticker-entries-wait
type: scenario
---
Given a delivery with ticker entries of a confirmed and an unconfirmed ticker, when the core imports it, then the confirmed ticker's entries are booked with its ISIN and the others wait, each an ingest problem saying so.