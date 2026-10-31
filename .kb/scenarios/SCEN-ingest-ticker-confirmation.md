---
title: The owner confirms a ticker's ISIN once
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:19.712Z'
id: SCEN-ingest-ticker-confirmation
type: scenario
---
Given tickers waiting with and without a proposal, when the owner confirms the proposal or types an ISIN through the TickerService, then the ISIN is stored once, every entry waiting for it is booked, a later upload books the ticker's rows at once, and another ISIN, an unknown ticker or a wrong check digit is refused.