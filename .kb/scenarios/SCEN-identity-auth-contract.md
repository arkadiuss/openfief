---
title: A view plugin calls the core's AuthService over Connect
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:34:54.847Z'
id: SCEN-identity-auth-contract
type: scenario
---
Given the core's plugin port, which only plugin pods reach (D-115), when a view plugin calls the AuthService over Connect with its own token, then the generated client reaches the generated handler and the core answers; the same procedure on the core's browser port is not found.