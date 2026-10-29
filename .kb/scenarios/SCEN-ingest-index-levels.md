---
title: A connector's index levels are translated by index name and stored with the instance as their source
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:42:22.413Z'
id: SCEN-ingest-index-levels
type: scenario
---
Given a delivery whose index_levels carry FAKE20's level 1522.60 on a date, when the core imports it, then the level becomes a valuation index level under the same name and is stored with the connector instance as its source; a level without a name, with a zero level, a malformed date or an exponent is refused like a bad price.