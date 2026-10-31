---
title: The NBP FX plugin's manifest and first-sync start
status: active
sourceFile: plugins/nbp-fx/main_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:46.106Z'
id: TEST-plugins-nbp-fx-main
type: test
verification_scope: unit
---
TestManifest parses plugins/nbp-fx/manifest.json, the image's signed label, with plugins.ParseManifestLabel and requires exactly the connector capability, the egress host api.nbp.pl on port 443, no Secrets and a config schema. TestStartDate checks a configured start, the default of the 93 days up to a fixed today, and a malformed date. TestToday checks the clock gives a date. A false pass would be a manifest granting more hosts; the exact comparison guards it.