---
title: A granted plugin gets an isolated pod, policy and identity
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:20.988Z'
id: SCEN-operator-workload-objects
type: scenario
---
Given the sample connector granted api.example.com:443 and its token Secret, when its workload is mapped, then its ServiceAccount mounts no API token, its pod mounts only the projected token for openfief-core and the Secret read-only, its NetworkPolicy allows ingress from the core pods on 9090 and egress to the core pods on 8080 and the proxy pods on 4751 only, and HTTPS_PROXY points at the proxy ClusterIP. Without granted egress, the proxy rule and proxy variables are absent.