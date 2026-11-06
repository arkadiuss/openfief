---
title: Open Positions asserts each stock or ETF ticker's summed volume
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:20.151Z'
id: SCEN-ingest-xtb-holdings-mapping
type: scenario
---
Given an xStation export with two open positions of one stock, one of an ETF and a CFD, when it is mapped in either mode, then each stock or ETF ticker's summed volume is a ticker assertion on the report's date, the CFD none; a ticker with a position that can't be read is an ingest problem and not asserted; an export without the report's date asserts nothing.