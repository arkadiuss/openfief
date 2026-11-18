---
title: A token granted one account reads that account alone
status: active
origin:
  kind: agent
  ref: the design session's answers on D-213, D-209, D-165 and D-199, relayed by the coordinator, 2026-11-18
  recorded_at: '2026-11-18T17:00:00.000Z'
id: SCEN-identity-grants-limit-tokens-under-policy
type: scenario
---
Given tokens granted the household, one account and nothing, and a view plugin granted a sleeve, then under the policy each views exactly the accounts its grants cover over the read API and, for an MCP token, through the tools, and changes none.