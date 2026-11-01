---
title: An xStation export is mapped in transactions mode
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:51.066Z'
id: SCEN-ingest-xtb-export-mapping
type: scenario
---
Given an uploaded xStation export of one account in transactions mode, when the xtb connector maps it, then its deposits and withdrawals become transactions, its stock and ETF buys, sells and dividends with their withheld tax become ticker entries, oldest first, named by the XTB account, with amounts rounded to the account currency's minor unit.