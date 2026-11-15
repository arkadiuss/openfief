---
title: At most the configured number of the Ask box's calls to a model run at once in a core process; a call past it waits up to the configured queue wait for a free place and is then refused as busy, recorded as an audit entry of class busy, a failed call frees its place and a caller that gives up while waiting is not recorded as a refusal; with no limit set no call waits.
status: open
firmness: default
semantic_text: At most the configured number of the Ask box's calls to a model run at once in a core process; a call past it waits up to the configured queue wait for a free place and is then refused as busy, recorded as an audit entry of class busy, a failed call frees its place and a caller that gives up while waiting is not recorded as a refusal; with no limit set no call waits.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:37:46.311Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 44fb9afac0b68e43a7da0375875c434447b9bb2b56912c34cae1452196cf44f7
semantic_inventory:
  - claim_key: CLAIM-37B66E3BF4B787D5
    claim_text: At most the configured number of the Ask box's calls to a model run at once in a core process
    role: descriptive
    status: missing
    span:
      start: 0
      end: 93
  - claim_key: CLAIM-70546193394E0DEA
    claim_text: a call past it waits up to the configured queue wait for a free place and is then refused as busy, recorded as an audit entry of class busy, a failed call frees its place and a caller that gives up while waiting is not recorded as a refusal
    role: descriptive
    status: missing
    span:
      start: 95
      end: 335
  - claim_key: CLAIM-38BD7B0D824CF025
    claim_text: with no limit set no call waits
    role: descriptive
    status: missing
    span:
      start: 337
      end: 368
logic_claims:
  - CLAIM-37B66E3BF4B787D5
  - CLAIM-70546193394E0DEA
  - CLAIM-38BD7B0D824CF025
id: REQ-models-ask-concurrent-calls
type: req
---
At most the configured number of the Ask box's calls to a model run at once in a core process; a call past it waits up to the configured queue wait for a free place and is then refused as busy, recorded as an audit entry of class busy, a failed call frees its place and a caller that gives up while waiting is not recorded as a refusal; with no limit set no call waits.

## Context

D-208's third guard also bounds the calls at once to 3. A short queue absorbs the ordinary burst, and a bounded one keeps the wait from becoming a cost of its own. The limit is per process, which is the demo's one replica.

## Source

> The design session's M5 corrections (2026-11-14): At most 3 model calls at once, a global semaphore in the router for the Ask box's role; queue briefly or refuse fast, chosen and explained (M5 brief, 2026-11-14).