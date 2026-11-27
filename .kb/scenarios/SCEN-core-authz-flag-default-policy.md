---
title: Without --authz=spicedb the core behaves as it did with the policy
status: closed
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:42.577Z'
id: SCEN-core-authz-flag-default-policy
type: scenario
---
Closed on 2026-11-22: SpiceDB became the default authorizer (D-164, the author's decision relayed by the design session), so a core with no --authz flag no longer runs the policy. SCEN-core-authz-flag-default-spicedb replaces this scenario. What it said: given the core started with no --authz flag or with --authz=policy, and with or without SpiceDB's settings in the environment, then the authorizer is identityapp.Policy, no connection to SpiceDB is made and the registry has no observer: the owner acts on every account and every other holder of read access views every account. Assumed nothing of SpiceDB is read for the policy.