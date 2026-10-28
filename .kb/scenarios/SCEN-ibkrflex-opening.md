---
title: An opening Flex statement seeds opening lots and opening cash on the go-live date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:18.794Z'
id: SCEN-ibkrflex-opening
type: scenario
---
Given a synthetic Flex statement ending 2026-03-06 with lot-level open positions of a stock (two lots opened 2024-01-15 and 2025-06-10 with their cost basis) and an ETF (two lots opened the same day without an originating transaction ID), ending cash of 1250.75 USD and -40.1 EUR, and one trade, when it is mapped in opening mode, then each lot becomes an opening lot dated 2026-03-06 with its open date, units and cost, keyed by ISIN, open date and IBKR's originating ID or ordinal; the cash becomes a deposit and a withdrawal; the trade is skipped; and an option lot is an unsupported problem.