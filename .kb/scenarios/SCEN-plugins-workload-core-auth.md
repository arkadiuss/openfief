---
title: Every plugin gets the signing keys and its instance name
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:29.369Z'
id: SCEN-plugins-workload-core-auth
type: scenario
---
Given a plugin sample granted a connector capability, when its workload is compiled and mapped to a pod, then the pod mounts the ConfigMap openfief-sa-jwks read-only at /var/run/secrets/openfief-sa-jwks and its environment sets OPENFIEF_INSTANCE to sample, OPENFIEF_CORE_SUBJECT to the core's ServiceAccount username and OPENFIEF_SA_JWKS_DIR to the mount, with or without egress; a topology without the core's ServiceAccount is refused.