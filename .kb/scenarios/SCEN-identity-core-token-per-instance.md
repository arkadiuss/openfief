---
title: The core presents a fresh, instance-scoped TokenRequest token to each plugin instance
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:36.332Z'
id: SCEN-identity-core-token-per-instance
type: scenario
---
Given the core's ServiceAccount openfief-core, when the core calls plugin instance ibkr-main, then it presents a token requested through TokenRequest with audience openfief-plugin:ibkr-main valid for 10 minutes, reused for that instance until 2 minutes before it expires; another instance gets a token with its own audience; a failed request is not cached and its error never holds a token.