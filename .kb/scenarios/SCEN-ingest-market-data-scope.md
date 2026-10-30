---
title: A market-data connector is asked to price what is held and what the model targets
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:19.526Z'
id: SCEN-ingest-market-data-scope
type: scenario
---
Given holdings of a stock and a treasury bond with terms in acc-1 and of an ETF and the stock in acc-2, and a model portfolio that is on with the stock and the FAKE20 index, when an instance bound to no account syncs, its request names the ETF and the stock once and FAKE20, in transactions mode; an instance bound to an account gets no instruments; a model that is off adds nothing.