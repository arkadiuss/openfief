---
title: XTB opening lots carry their open price, not a cost
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T16:48:16.918Z'
id: SCEN-ingest-xtb-opening-lot-trading-currency
type: scenario
---
Given a current xStation export in opening mode, when the xtb connector maps it, then each open stock or ETF position becomes an opening lot on the report's date with its volume, open date and open price and no cost, whatever currency it trades in, and a position on a market of no known exchange is an ingest problem.