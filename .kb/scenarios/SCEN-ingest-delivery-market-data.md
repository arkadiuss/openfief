---
title: A plugin writes prices and index levels as exact contract messages
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:18.578Z'
id: SCEN-ingest-delivery-market-data
type: scenario
---
Given a synthetic instrument IETEST000026 with venue symbol TSTB and a close of 84.1250 GBP on 2025-01-03, when the plugin writes it, then the contract price names that ISIN and symbol, the date 2025-01-03 and the amount 84.125 in GBP; given FAKE20's level 1522.60 on 2001-01-03, then the contract index level names FAKE20, that date and the decimal string 1522.6, with no float anywhere.