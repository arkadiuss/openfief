---
title: A quote is asked of the highest-ranked declared connector that takes quotes and of no other
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-ingest-market-quote-source
type: scenario
---
Given several connector instances, only some declared and approved as market-data connectors, when a quote is looked up, then the owner's highest-ranked one is asked, one that takes no quotes is passed over for the next, one that is down or finds nothing is the answer, one not declared is never asked, and with none left the lookup finds no source. Assumes instances and declarations as the cluster reports them.
