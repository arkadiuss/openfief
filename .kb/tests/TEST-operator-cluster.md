---
title: Core and egress proxy addresses come only from usable Services
status: active
sourceFile: internal/plugins/operator/cluster_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:15.921Z'
id: TEST-operator-cluster
type: test
---
Table tests build synthetic core and egress proxy Services and call ClusterFromServices: ClusterIPs (IPv4 and IPv6), ports and numeric or defaulted target ports become the topology and the selectors become NetworkPolicy peers; a headless Service, a named target port, a UDP port, a missing port, a missing selector and selectors that select one or every plugin's pods are refused with their sentinel errors. A false pass would be a core selector that matches a plugin; two cases cover it.