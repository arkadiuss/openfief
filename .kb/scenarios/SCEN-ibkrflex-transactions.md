---
title: A Flex statement in transactions mode books trades and cash transactions under IBKR's own IDs
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:20.884Z'
id: SCEN-ibkrflex-transactions
type: scenario
---
Given a synthetic Flex statement with stock and ETF executions, EUR.USD currency trades, deposits, a withdrawal, a market data fee and dividends, when it is mapped in transactions mode, then each stock execution becomes a buy or sell keyed by its tradeID with commission and taxes as the fee, each currency trade an FX conversion of quote against base currency, and each cash transaction a deposit, withdrawal, fee or dividend keyed by its transactionID, dated on its calendar date as IBKR reports it; order rows, closed-lot rows, cancellations, options and broker interest become ingest problems.