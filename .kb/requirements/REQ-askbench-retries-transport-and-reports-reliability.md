---
title: The Ask box bench retries a call that failed with no answer, a timeout, a 5xx or a bad reply once, retries a 429 once after its Retry-After inside the chat client, and never retries a 4xx or an answer; it counts the retry as a call and marks it in the run detail, counts reliability per question and not per call (a question is failed only when, after its retry, it still has no usable answer because of a host fault, however many calls it took), reports each model@host's failed-question rate beside its failed-call counts and fails a host whose rate is over five percent of its questions for the run whatever its verdict.
status: open
tags:
  - askbench
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'Brief of the Ask bench 4xx and per-question reliability change: the design delegate''s decisions on D-207 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T20:00:00.000Z'
semantic_text: The Ask box bench retries a call that failed with no answer, a timeout, a 5xx or a bad reply once, retries a 429 once after its Retry-After inside the chat client, and never retries a 4xx or an answer; it counts the retry as a call and marks it in the run detail, counts reliability per question and not per call (a question is failed only when, after its retry, it still has no usable answer because of a host fault, however many calls it took), reports each model@host's failed-question rate beside its failed-call counts and fails a host whose rate is over five percent of its questions for the run whatever its verdict.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4a56db447219b1f3c66d0b97faecd0a980e5a11c8a4c5d2866c3691a98046706
semantic_inventory:
  - claim_key: CLAIM-AD30E88E91F6E424
    claim_text: The Ask box bench retries a call that failed with no answer, a timeout, a 5xx or a bad reply once, retries a 429 once after its Retry-After inside the chat client, and never retries a 4xx or an answer
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 200
  - claim_key: CLAIM-E611530DDF87C828
    claim_text: it counts the retry as a call and marks it in the run detail, counts reliability per question and not per call (a question is failed only when, after its retry, it still has no usable answer because of a host fault, however many calls it took), reports each model@host's failed-question rate beside its failed-call counts and fails a host whose rate is over five percent of its questions for the run whatever its verdict
    role: rationale
    status: nonlogical
    span:
      start: 202
      end: 622
logic_claims:
  - CLAIM-AD30E88E91F6E424
id: REQ-askbench-retries-transport-and-reports-reliability
type: req
---
The Ask box bench retries a call that failed with no answer, a timeout, a 5xx or a bad reply once, retries a 429 once after its Retry-After inside the chat client, and never retries a 4xx or an answer; it counts the retry as a call and marks it in the run detail, counts reliability per question and not per call (a question is failed only when, after its retry, it still has no usable answer because of a host fault, however many calls it took), reports each model@host's failed-question rate beside its failed-call counts and fails a host whose rate is over five percent of its questions for the run whatever its verdict.

## Context

Decided in the AB6 design session (2026-11-14) and the coordinator's D-221 addition: a call that fails on the network says nothing of the model, a host that fails too often should not serve the box whatever its answers, and a model should not join or leave the demo route on a single run. Changed on 2026-11-15 by the design session (D-207): a 429 is the host rate-limiting its callers, so it is retried once honouring Retry-After and, when it persists, counts against the host as the one 4xx that does.

## Source

> The AB6 brief and the coordinator's D-221 addition, 2026-11-14, and the design session's 429 decision, 2026-11-15.
