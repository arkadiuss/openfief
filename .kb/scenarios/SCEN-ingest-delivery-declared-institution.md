---
title: A connector's delivery naming an undeclared institution is refused whole
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T22:19:43.334Z'
id: SCEN-ingest-delivery-declared-institution
type: scenario
---
Given a connector instance named aaa whose signed manifest declares only the institution aaa-broker, or none, when its delivery names eodhd, then the core refuses the whole delivery with ErrUndeclaredInstitution: no transaction is imported, no price is stored and aaa is not recorded as a market-data source, so it can't take EODHD's default place in the ranking; a delivery naming a declared institution is imported as before.