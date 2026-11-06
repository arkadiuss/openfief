---
title: A connector's credentials exist only in its Kubernetes Secret
status: active
tags:
  - acceptance:p10-14
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:41.443Z'
id: SCEN-e2e-credentials-only-in-secrets
type: scenario
---
Given the synthetic connector instance e2e-books with a random synthetic credential in its Secret e2e-books--credentials, after a sync in which the connector reads the credential and books a deposit, when every ConfigMap in the cluster, every pod spec with its environment, command and arguments, every Plugin and PluginGrant, and a pg_dumpall of the database are searched for the credential, then it is found in none of them. Runs on both kind and k3d.