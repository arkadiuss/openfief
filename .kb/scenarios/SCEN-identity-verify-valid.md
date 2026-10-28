---
title: A stored, unexpired, unrevoked read token is accepted
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:10.002Z'
id: SCEN-identity-verify-valid
type: scenario
---
Given an issued read token, when a script presents its secret before the expiry for the read scope, then Verify finds the token by its ID, compares the key's hash in constant time and accepts it.