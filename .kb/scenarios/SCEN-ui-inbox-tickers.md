---
title: Tickers are confirmed in the Inbox
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:06.352Z'
id: SCEN-ui-inbox-tickers
type: scenario
---
Given tickers waiting for their ISIN, when the owner opens the Inbox, then each ticker's item has the proposal filled in or an empty field, and a trading-currency select when its exchange quotes in more than one, required without a proposed currency; a confirmation reaches the core trimmed and in upper case and the Inbox says how many rows it booked.