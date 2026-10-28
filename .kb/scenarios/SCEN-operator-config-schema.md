---
title: A config its schema rejects is not deployed
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:25.803Z'
id: SCEN-operator-config-schema
type: scenario
---
Given a manifest schema that requires a string interval, when the Plugin's config sets interval to 24, then the config is rejected and that version is not deployed. Given a schema that references https://schemas.example.com/config.json, then the schema is refused without any network call.