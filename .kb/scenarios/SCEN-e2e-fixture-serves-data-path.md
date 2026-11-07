---
title: The fixture serves the data path Flex statements by query ID, its NBP rates and EODHD listings
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:41.525Z'
id: SCEN-e2e-fixture-serves-data-path
type: scenario
---
Given the fixture handler, when the IBKR Flex client asks for each data path query ID, the NBP client for the data path table A and the EODHD client for its instruments, then each statement maps to the opening, week and corrected transactions, the rates rise each weekday for GBP, CHF and NOK only, and the listings and closes are the synthetic ones.