---
title: Discovery reaches an instance as a model provider only while it runs as one
status: active
sourceFile: internal/ingest/kube/modelprovider_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:07.879Z'
id: TEST-ingest-kube-model-provider
type: test
---
Reads Plugin resources and Services through fake clients: a ready model-provider instance has an address; connectors, views, instances not ready and absent ones have none.