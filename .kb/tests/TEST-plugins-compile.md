---
title: Compiled workloads allow only the core and the proxy, mount only granted Secrets and carry the core audience token
status: active
sourceFile: internal/plugins/compile_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:27:58.203Z'
id: TEST-plugins-compile
type: test
---
Table tests compile synthetic effective permissions against a topology with documentation-range ClusterIPs and compare the whole Workload: ingress only from the core on the plugin port, egress to the core's pod port plus the proxy's pod port only when an egress host is granted, HTTPS_PROXY and NO_PROXY only then (with IPv6 addresses bracketed), granted Secrets mounted read-only, ServiceAccount plugin-<name> in openfief with no API token, and a projected token with audience openfief-core. Separate tests check that no rule allows DNS port 53, that reordered inputs compile identically, and that invalid topologies and empty permissions fail. A false pass would be a proxy route for a plugin without egress; the no-egress case asserts its absence.