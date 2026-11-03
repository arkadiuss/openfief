---
title: Every service's failures are logged by their kind only
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T03:24:19.045Z'
id: SCEN-core-logs-carry-no-values
type: scenario
---
Given failures whose text carries an amount, an ISIN and an account number, when a Connect service of the ledger, valuation, views, identity, portfolio or attention context, the built-in UI or the core's attention sweep logs one, then the line names only the known errors in its chain or its innermost error's type, and none of those values (GR-12).