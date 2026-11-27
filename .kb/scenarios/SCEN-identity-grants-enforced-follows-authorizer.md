---
title: The core reports grants as enforced under SpiceDB and under the policy
status: active
origin:
  kind: agent
  ref: the owner's items for alpha.2, from the reviews of main 0bfeb62
  recorded_at: '2026-11-27T08:26:26.250Z'
id: SCEN-identity-grants-enforced-follows-authorizer
type: scenario
---
Given the authorizer is SpiceDB, or the policy over the grants table (--authz=policy), the lists of script tokens, view plugins and people say grants are enforced; given an authorizer that lets every holder of read access view every account, or none given, they say they are not.