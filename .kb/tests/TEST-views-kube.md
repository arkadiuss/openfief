---
title: Installed plugins are read from the operator's Plugin resources through a watched cache
status: active
sourceFile: internal/views/kube/plugins_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:26.196Z'
id: TEST-views-kube
type: test
---
Runs the adapter over client-go's fake dynamic client: before Start it fails with ErrPluginsUnknown; after Start it returns every Plugin in the namespace, a deleted one included as deleting, with its image, requested and effective capabilities, Ready and SignatureVerified conditions and status.signature; a Plugin added later appears once the watch delivers it. A contract case converts a v1alpha1.Plugin the operator writes into unstructured data, so a renamed status field fails the test.