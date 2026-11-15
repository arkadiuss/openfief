---
title: The Ask box bench retries a call that failed with no answer, a timeout or a 5xx once and never a 4xx or an answer, counts the retry as a call and marks it in the run detail, reports each model@host's transport-failure rate and fails a host whose rate is over five percent for the run whatever its verdict.
status: open
tags:
  - askbench
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:46.891Z'
semantic_text: The Ask box bench retries a call that failed with no answer, a timeout or a 5xx once and never a 4xx or an answer, counts the retry as a call and marks it in the run detail, reports each model@host's transport-failure rate and fails a host whose rate is over five percent for the run whatever its verdict.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 689cf73ae45d857ee6fc867f47008e2c65fd2618383e7b6884eb6e047157945e
semantic_inventory:
  - claim_key: CLAIM-3ADF8CC9B126D9A3
    claim_text: The Ask box bench retries a call that failed with no answer, a timeout or a 5xx once and never a 4xx or an answer, counts the retry as a call and marks it in the run detail, reports each model@host's transport-failure rate and fails a host whose rate is over five percent for the run whatever its verdict
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 304
logic_claims:
  - CLAIM-3ADF8CC9B126D9A3
id: REQ-askbench-retries-transport-and-reports-reliability
type: req
---
The Ask box bench retries a call that failed with no answer, a timeout or a 5xx once and never a 4xx or an answer, counts the retry as a call and marks it in the run detail, reports each model@host's transport-failure rate and fails a host whose rate is over five percent for the run whatever its verdict.

## Context

Decided in the AB6 design session (2026-11-14) and the coordinator's D-221 addition: a call that fails on the network says nothing of the model, a host that fails too often should not serve the box whatever its answers, and a model should not join or leave the demo route on a single run.

## Source

> The AB6 brief and the coordinator's D-221 addition, 2026-11-14: Decided in the AB6 design session (2026-11-14) and the coordinator's D-221 addition: a call that fails on the network says nothing of the model, a host that fails too often should not serve the box whatever its answers, and a model should not join or leave the demo route on a single run.
