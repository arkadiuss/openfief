---
title: The models store keeps the settings in the models schema
status: active
sourceFile: internal/modelrouter/postgres/store_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:02.018Z'
id: TEST-models-postgres-store
type: test
---
Runs the store on a real database: off until saved, round trips, replacing and clearing routes and tier lists, and turning the switch off again.