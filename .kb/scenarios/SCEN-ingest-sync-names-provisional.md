---
title: A sync names provisional instruments and the connector answers their registry ISIN
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:27.348Z'
id: SCEN-ingest-sync-names-provisional
type: scenario
---
Given the accounts bound to a treasury-bonds instance post to instruments under a provisional ISIN, held or not, when the core syncs the instance, then the request names each once, and the connector answers the registry ISIN of each whose series the ministry's data now lists, nothing for the rest, and nothing while the data can't be read.