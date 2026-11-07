---
title: Attribution crosses the connector contract
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:10.911Z'
id: SCEN-ingest-bond-terms-attribution-contract
type: scenario
---
Given terms read from published data, when the connector delivers them and the core hands them back in a sync, then the attribution travels with them, and the core refuses one that lacks a part the licence asks for or whose dates don't parse.