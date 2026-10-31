---
title: The owner confirms tickers in the UI
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:23.462Z'
id: SCEN-ui-tickers-to-confirm
type: scenario
---
Given tickers waiting for their ISIN, when the owner opens the tickers page, then each is listed with its proposal to confirm or a field to type the ISIN, a confirmation is handed to the core, and an upload whose rows wait links the page.