---
title: The fixture serves the treasury-bonds connector's sources in their real formats
status: active
origin:
  kind: agent
  recorded_at: '2026-11-08T00:35:35.518Z'
id: SCEN-e2e-fixture-serves-bond-data
type: scenario
---
Given the institution fixture, when the treasury-bonds connector's ministry, GUS and NBP clients read it, then the ministry's data is withheld until the suite publishes it, and then the gov.pl page and dane.gov.pl's resource list both lead to a synthetic .xls whose sheets list each series' registry ISIN, first-period rate and margin; GUS answers the demo dataset's CPI and NBP's archive synthetic reference rate changes.