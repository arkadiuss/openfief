---
title: The EODHD plugin's manifest asks for exactly eodhd.com and the API key Secret, and its config is checked
status: active
sourceFile: plugins/eodhd/main_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:03.234Z'
id: TEST-plugins-eodhd-main
type: test
verification_scope: integration
---
TestManifest parses manifest.json with ParseManifestLabel and requires exactly the connector capability, egress to eodhd.com on 443 and the eodhd-credentials Secret at /etc/openfief/eodhd, plus a config schema. TestConfigParse accepts an empty config and one with instruments and indices, and refuses a non-date start, an ISIN with a wrong check digit, an unsupported exchange, an ISIN or index name twice and an index without a ticker. TestNewConnector checks the wiring to EODHD's client, the mounted Secret and the clock; TestToday that the clock gives a date.