---
title: The Ask box bench asks three questions at once unless -parallel says otherwise, because D-208 caps the box at three model calls at once and a question makes one call at a time.
status: open
origin:
  kind: agent
  ref: 'Design session decision: bench parallel 3 (D-208) and the performance tool''s description (D-203), 2026-11-15'
  recorded_at: '2026-11-15T10:09:50.502Z'
id: REQ-askbench-parallel-matches-the-box-cap
type: req
tags:
  - askbench
semantic_text: The Ask box bench asks three questions at once unless -parallel says otherwise, because D-208 caps the box at three model calls at once and a question makes one call at a time.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
required_evidence: unverified
semantic_source_hash: ba40836be42d657bdb856b73bde7c815283555c7d3cdeb8996ff79a41d33e88c
semantic_inventory:
  - claim_key: CLAIM-E3CFD8619AF961D0
    claim_text: The Ask box bench asks three questions at once unless -parallel says otherwise, because D-208 caps the box at three model calls at once and a question makes one call at a time
    role: rationale
    status: nonlogical
    span:
      start: 0
      end: 175
---
The Ask box bench asks three questions at once unless -parallel says otherwise, because D-208 caps the box at three model calls at once and a question makes one call at a time.

## Context

A bench run at four at once tested a load the box refuses, and raised rate limits (429) that no visitor of the demo could cause.

## Source

Design session decision, 2026-11-15: bench default -parallel becomes 3 (D-208's cap of calls at once).
