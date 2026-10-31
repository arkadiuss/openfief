---
title: Running connector instances and their addresses are discovered from Plugins and Services and watched
status: active
sourceFile: internal/ingest/kube/discovery_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:31.878Z'
id: TEST-ingest-kube
type: test
verification_scope: unit
---
Runs Discovery over the dynamic fake client with Plugin resources and the fake clientset with Services. Before Start it fails with ErrNotStarted; after Start only the Ready, not deleting plugin with connector among its effective capabilities and a Service is listed, a view plugin, a not-ready one, a deleting one and one without a Service are not; BaseURL returns http://ClusterIP:port and ErrNoAddress without a Service; creating the missing Service makes the watch list it. ServiceAddress handles IPv4, IPv6, no ClusterIP, headless and a missing connect port, and the label and port names match the operator's.