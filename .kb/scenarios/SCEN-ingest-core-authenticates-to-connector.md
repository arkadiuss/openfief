---
title: The core presents a fresh token for the instance on every call
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:57.284Z'
id: SCEN-ingest-core-authenticates-to-connector
type: scenario
---
Given connector instance xtb-main, when the core calls its Sync or ImportFile, then the call carries Authorization Bearer with a token the core's token issuer made for xtb-main just before the call; when the issuer fails or the instance has no address, the core makes no call.