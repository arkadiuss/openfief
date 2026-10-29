---
title: The read API lists what the onboarding forms need
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:39.470Z'
id: SCEN-views-onboarding-lists
type: scenario
---
Given the built-in kinds, a binding of ibkr-main to acc-1 and two broker accounts of ibkr-main, when a client with a valid token asks GET /v1/sleeve-kinds, /v1/connector-bindings and /v1/broker-accounts?unmapped=true, then it gets the six kinds with their default roles, the binding and only the unmapped broker account with its institution and last seen date; a sleeve carries its role override.