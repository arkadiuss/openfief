---
title: A plugin built on the SDK finds the core, its token, its port and its config
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:58.758Z'
id: SCEN-sdk-go-plugin-environment
type: scenario
---
Given the operator set OPENFIEF_CORE_ADDRESS, OPENFIEF_TOKEN_FILE, OPENFIEF_LISTEN_PORT and OPENFIEF_CONFIG in the plugin container, when the plugin reads its environment through the SDK, then it calls the core at that ClusterIP and port, reads its token from that file, listens on that port and decodes that JSON; a missing or malformed core address or port is an error rather than a guess.