---
title: 'A platform credential''s Secret that exists is kept as it is: an owner''s own, or the one an earlier run made'
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T07:41:53.564Z'
id: SCEN-identity-credentials-owner-fed-kept
type: scenario
---
Given the owner made openfief-spicedb-key, or both Secrets, from their own tooling (D-29), or an earlier run made them, when core init-credentials runs, then each creation answers already-exists, the Secret is neither read, changed nor rotated, the Secrets that are missing are still created, the owner's values are unchanged, and the log names each Secret as kept. The only calls to the API server are creates of Secrets: never get, list, watch, update, patch or delete.