---
title: Only a plugin with egress gets the egress-audience token
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:37:31.350Z'
id: SCEN-plugins-egress-token-only-with-egress
type: scenario
---
Given plugin sample granted api.example.com:443, when its workload is compiled and deployed, then its pod has the core token (audience openfief-core, OPENFIEF_TOKEN_FILE=/var/run/secrets/openfief/token) and a second projected token with audience openfief-egress at /var/run/secrets/openfief-egress/token, named by OPENFIEF_EGRESS_TOKEN_FILE, both expiring after 3600 s and mounted read-only. Given a plugin granted no egress host, its pod has only the core token and no OPENFIEF_EGRESS_TOKEN_FILE.