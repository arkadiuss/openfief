---
title: Bond lots booked before the ministry's data lists them are re-keyed to the registry ISIN, and terms show their source
status: active
tags:
  - acceptance:p10-5
origin:
  kind: agent
  recorded_at: '2026-11-08T00:35:39.291Z'
id: SCEN-e2e-bond-registry-isin-and-terms
type: scenario
---
Given the fixture withholds the Ministry of Finance's data while the owner enters a TOS0329 and an EDO0534 lot with their terms, so both are booked under provisional ISINs, when the fixture publishes the data and a sync runs, then both lots are under the registry ISINs the spreadsheet's Kod ISIN column lists, each buy superseded by a re-keyed version; a COI1029 lot entered with its terms left to the connector is booked under its registry ISIN, and the read API shows its published rates with the ministry's attribution and the fee table's fee, and the other two series' entered terms under their registry ISINs.