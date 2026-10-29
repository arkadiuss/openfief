---
title: The operator reaches a granted in-cluster service's pods by NetworkPolicy and hands the plugin its ClusterIP
status: active
sourceFile: internal/plugins/operator/cluster_service_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:01.117Z'
id: TEST-operator-cluster-service
type: test
---
AddClusterService refuses headless, selector-less, portless and named-targetPort Services. On envtest, a Service in namespace ghostfolio and a Plugin and grant naming it give a NetworkPolicy rule with a namespace selector and the Service's selector on pod port 3000, OPENFIEF_CLUSTER_SERVICE_ADDRESS with the assigned ClusterIP, no proxy and no allow-list entry; without approval it waits; a missing Service runs nothing and is retried.