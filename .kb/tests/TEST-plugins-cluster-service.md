---
title: An in-cluster service is parsed, validated, approved exactly and compiled to a direct rule outside the proxy
status: active
sourceFile: internal/plugins/cluster_service_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T15:09:58.394Z'
id: TEST-plugins-cluster-service
type: test
---
Table tests: ParseClusterService accepts namespace/service:port and refuses the openfief namespace, bad names and ports; a Plugin names one only when its manifest asks; NeedsApproval and Effective keep it only when the grant approves exactly that service and port; Compile adds a rule to the service's pods on the pod port, its ClusterIP in OPENFIEF_CLUSTER_SERVICE_ADDRESS and NO_PROXY, never an allow-list entry, and fails without its address.