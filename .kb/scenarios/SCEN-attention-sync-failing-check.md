---
title: A failing sync's item follows its run of failures
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T01:26:35.812Z'
id: SCEN-attention-sync-failing-check
type: scenario
---
Given a connector instance whose sync keeps failing, when the attention sweep runs, then one item of kind connector-sync-failing is open for the instance, its detail when the failures began; the owner's acknowledgement holds while the same run goes on; the next success resolves it; and a new run of failures reopens it.