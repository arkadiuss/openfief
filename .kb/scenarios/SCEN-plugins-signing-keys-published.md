---
title: The operator publishes the cluster's public signing keys for plugins
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:27.465Z'
id: SCEN-plugins-signing-keys-published
type: scenario
---
Given a cluster with default RBAC, when the operator starts, then the ConfigMap openfief-sa-jwks holds the API server's issuer and its public signing keys, read as the operator's ServiceAccount; when the ConfigMap is deleted it is published again; when the API server can't be read, the published keys stay; and a key set with private or symmetric key material is never published.