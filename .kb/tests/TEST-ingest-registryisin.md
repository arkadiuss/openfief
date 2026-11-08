---
title: Registry ISINs are validated and ordered; provisional instruments are found in transactions
status: active
sourceFile: internal/ingest/registryisin_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:13.099Z'
id: TEST-ingest-registryisin
type: test
---
Table tests of ingest.RegistryISIN.Validate over a provisional series re-keyed to a synthetic registry ISIN and every refusal (a registry ISIN already, another series' provisional ISIN, a listed instrument, a non-Polish or provisional target, no ISIN, no instrument), of CheckRegistryISINs ordering by instrument and refusing an instrument named twice, and of ProvisionalInstruments listing each provisional instrument a holding posting names once, in order, and nothing else. A false pass would let a connector re-key an instrument that is not a provisional bond series.