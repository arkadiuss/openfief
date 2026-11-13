---
title: The database, egress proxy, operator and SpiceDB policies admit only their callers
status: active
sourceFile: scripts/tests/network-policies.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:16.354Z'
id: TEST-deploy-network-policies
type: test
---
The script test builds deploy/core, deploy/egress-proxy, deploy/operator, deploy/demo-data and deploy/spicedb with kustomize and checks the built NetworkPolicies with gojq: the database's admits exactly the core on 5432, its own instances on 5432 and 8000 and CloudNativePG's operator on 8000, and no plugin; the demo loader brings its own admission; the proxy admits plugin pods on 4750 alone and its egress leaves every private, shared, link-local and loopback range out; the operator admits nothing; SpiceDB admits the core on 50051 and nothing else and reaches the database on 5432 and DNS only, the database admits SpiceDB on 5432 and the core reaches it on 50051 through policies of the spicedb base while the core's own policy names it nowhere, no policy of the base names a plugin, the credentials Job reaches the API server and DNS only and nothing reaches it, and SpiceDB's ServiceAccount has no rights or token while only the Job may create Secrets. Any extra rule fails a case.