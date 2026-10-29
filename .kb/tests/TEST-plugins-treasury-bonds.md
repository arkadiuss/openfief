---
title: The treasury-bonds plugin maps the owner's lots as the core's connector and asks for nothing else
status: active
sourceFile: plugins/treasury-bonds/connector_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:50.547Z'
id: TEST-plugins-treasury-bonds
type: test
---
Serves the plugin's connector with the Go SDK behind a fake core-token verifier and calls it as the core: ImportBondLots answers the exact delivery of a synthetic EDO0536 lot (deposit, buy, terms under PLEDO0536000); a bad series, mismatched or future date, half a bond, no fee, a rate with a percent sign and a duplicate lot are INVALID_ARGUMENT; no core token is UNAUTHENTICATED; Sync delivers nothing and keeps the watermark. main_test.go checks the manifest asks only for the connector capability and dates today in Warsaw.