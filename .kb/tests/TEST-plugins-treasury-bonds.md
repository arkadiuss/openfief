---
title: The treasury-bonds plugin maps lots and syncs index readings for the core, reaching only GUS and NBP
status: active
sourceFile: plugins/treasury-bonds/connector_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:50.547Z'
id: TEST-plugins-treasury-bonds
type: test
---
Serves the plugin's connector with the Go SDK behind a fake core-token verifier and calls it as the core: ImportBondLots answers the exact delivery of a synthetic EDO0536 lot (deposit, buy, terms under PLEDO0536000); a bad series, mismatched or future date, half a bond, no fee, a rate with a percent sign and a duplicate lot are INVALID_ARGUMENT; no core token is UNAUTHENTICATED. Sync, with fake GUS and NBP sources, delivers the exact CPI and reference-rate readings dated on each month's first day with their watermark, then nothing new; a foreign watermark is INVALID_ARGUMENT and an outage UNAVAILABLE. main_test.go checks the manifest asks for the connector capability and egress to api-sdp.stat.gov.pl and static.nbp.pl only, the start config and today in Warsaw.