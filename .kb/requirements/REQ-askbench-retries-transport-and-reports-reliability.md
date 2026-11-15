---
title: The Ask box bench retries a call that failed with no answer, a timeout or a 5xx once, retries a 429 once after its Retry-After inside the chat client, and never retries any other 4xx or an answer; it counts the retry as a call and marks it in the run detail, counts a 429 that stays a 429 as a transport failure, reports each model@host's transport-failure rate and fails a host whose rate is over five percent for the run whatever its verdict.
status: open
tags:
  - askbench
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:46.891Z'
semantic_text: The Ask box bench retries a call that failed with no answer, a timeout or a 5xx once, retries a 429 once after its Retry-After inside the chat client, and never retries any other 4xx or an answer; it counts the retry as a call and marks it in the run detail, counts a 429 that stays a 429 as a transport failure, reports each model@host's transport-failure rate and fails a host whose rate is over five percent for the run whatever its verdict.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 99f7f219301d278c7d53eb0eee2aed284e5915d78b2c5b2cd133e77ddd67f81d
semantic_inventory:
  - claim_key: CLAIM-A73EF4164D32FE47
    claim_text: The Ask box bench retries a call that failed with no answer, a timeout or a 5xx once, retries a 429 once after its Retry-After inside the chat client, and never retries any other 4xx or an answer
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 195
  - claim_key: CLAIM-483A68C39331334E
    claim_text: it counts the retry as a call and marks it in the run detail, counts a 429 that stays a 429 as a transport failure, reports each model@host's transport-failure rate and fails a host whose rate is over five percent for the run whatever its verdict
    role: descriptive
    status: ontology_gap
    span:
      start: 197
      end: 443
logic_claims:
  - CLAIM-A73EF4164D32FE47
  - CLAIM-483A68C39331334E
id: REQ-askbench-retries-transport-and-reports-reliability
type: req
---
The Ask box bench retries a call that failed with no answer, a timeout or a 5xx once, retries a 429 once after its Retry-After inside the chat client, and never retries any other 4xx or an answer; it counts the retry as a call and marks it in the run detail, counts a 429 that stays a 429 as a transport failure, reports each model@host's transport-failure rate and fails a host whose rate is over five percent for the run whatever its verdict.

## Context

Decided in the AB6 design session (2026-11-14) and the coordinator's D-221 addition: a call that fails on the network says nothing of the model, a host that fails too often should not serve the box whatever its answers, and a model should not join or leave the demo route on a single run. Changed on 2026-11-15 by the design session (D-207): a 429 is the host rate-limiting its callers, so it is retried once honouring Retry-After and, when it persists, counts against the host as the one 4xx that does.

## Source

> The AB6 brief and the coordinator's D-221 addition, 2026-11-14, and the design session's 429 decision, 2026-11-15.
