---
title: A compiled workload has its own identity and only granted Secrets
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:09.832Z'
id: SCEN-plugins-workload-identity-secrets
type: scenario
---
Given a connector granted the Secrets sample-cert and sample-token, when its workload is compiled, then it runs in namespace openfief as ServiceAccount plugin-sample with no Kubernetes API token and no role, presents a projected token with audience openfief-core mounted under /var/run/secrets/openfief, and mounts exactly those two Secrets read-only at their requested paths, sorted by name.