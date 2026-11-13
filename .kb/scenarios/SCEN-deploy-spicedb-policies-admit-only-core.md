---
title: SpiceDB admits only the core, reaches only the database and DNS, and the database and the core admit it through policies of its own base
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:16.784Z'
id: SCEN-deploy-spicedb-policies-admit-only-core
type: scenario
---
Given deploy/spicedb and deploy/core built with kustomize, then SpiceDB's NetworkPolicy admits the core's pods on 50051 and nothing else, plugins included, and lets SpiceDB out to the openfief-db cluster on 5432 and to DNS only; the core's egress to SpiceDB's pods on 50051 and the database's ingress from SpiceDB's pods on 5432 are policies of the spicedb base, so the core's and the database's own policies don't name SpiceDB; the credentials Job's pod reaches the API server and DNS only and nothing reaches it; and SpiceDB's ServiceAccount has no API rights and no token while only the Job may create Secrets.