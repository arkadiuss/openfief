---
title: The NBP FX plugin syncs table A rates after its watermark
status: active
sourceFile: plugins/nbp-fx/connector_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:43.609Z'
id: TEST-plugins-nbp-fx
type: test
verification_scope: integration
---
Table tests call the connector's Sync with a fake rate provider that publishes synthetic USD rates on chosen days and records the ranges asked. A first sync asks from the configured start to today; a later one from the day after the watermark; the watermark moves to the latest delivered rate and stays when nothing is published; a sync already at today asks nothing; a malformed watermark is InvalidArgument and a failing provider Unavailable. TestServedConnector serves the plugin with the SDK and checks Sync answers the core, refuses another caller and that ImportFile is Unimplemented. A false pass would be a watermark that skips days; the exact ranges are compared.