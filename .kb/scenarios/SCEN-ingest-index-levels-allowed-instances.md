---
title: Index levels of an allowed instance are translated and stored with the instance as source
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T22:24:58.999Z'
id: SCEN-ingest-index-levels-allowed-instances
type: scenario
---
Given a delivery with market index levels, when the core translates it, then each level is keyed by the index name the connector sends, one without a name, a date or a positive level refuses the delivery, and the levels of an instance that may deliver them are stored with that instance as their source while any other instance's are dropped (D-51, D-179).