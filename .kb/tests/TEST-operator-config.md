---
title: Plugin configs are checked against their manifest's JSON Schema without loading other documents
status: active
sourceFile: internal/plugins/operator/config_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:18.489Z'
id: TEST-operator-config
type: test
---
Table tests call ValidateConfig with a synthetic schema requiring a string interval: a matching config passes, no schema accepts anything, a wrong type, an unknown field, a missing config (checked as an empty object) and a config that is not JSON are rejected, schemas that are not JSON or not schemas are invalid, and schemas that reference an https or file URL are refused, while a local $defs reference works. A false pass would be a remote reference being fetched; the test runs offline and expects ErrInvalidConfigSchema.