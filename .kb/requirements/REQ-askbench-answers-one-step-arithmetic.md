---
title: '''''''The Ask box bench accepts a number in an answer that is one step of two numbers in the cards shown, their difference or their sum or, with a percent sign, the percentage change of one to the other, computed exactly and rounded half-up to the places the answer writes (at most two), and nothing else: no chain of operations and no number from outside the cards.'''''''
status: open
firmness: default
required_evidence: unverified
tags:
  - askbench
origin:
  kind: agent
  ref: Design session decisions on A19 as a false-premise case and one-step arithmetic in the numbers check, 2026-11-14
  recorded_at: '2026-11-14T22:23:37.946Z'
id: REQ-askbench-answers-one-step-arithmetic
type: req
semantic_text: '''''''The Ask box bench accepts a number in an answer that is one step of two numbers in the cards shown, their difference or their sum or, with a percent sign, the percentage change of one to the other, computed exactly and rounded half-up to the places the answer writes (at most two), and nothing else: no chain of operations and no number from outside the cards.'''''''
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8676940beba519f90fd7d3309a7902b798b4368e0e504604fad48c14131de704
semantic_inventory:
  - claim_key: CLAIM-DE5FC57FDA919373
    claim_text: '''''''The Ask box bench accepts a number in an answer that is one step of two numbers in the cards shown, their difference or their sum or, with a percent sign, the percentage change of one to the other, computed exactly and rounded half-up to the places the answer writes (at most two), and nothing else: no chain of operations and no number from outside the cards.'''''''
    role: descriptive
    status: missing
    span:
      start: 0
      end: 366
logic_claims:
  - CLAIM-DE5FC57FDA919373
---
'''The Ask box bench accepts a number in an answer that is one step of two numbers in the cards shown, their difference or their sum or, with a percent sign, the percentage change of one to the other, computed exactly and rounded half-up to the places the answer writes (at most two), and nothing else: no chain of operations and no number from outside the cards.'''

## Context

The live bench's gpt-oss-120b answers wrote the change between two shown values (an increase of 1,966.85 PLN) and failed the numbers check for a number no single card holds; and A19 asked why a value dropped when it rose. Decided in the design session of 2026-11-14.

## Source

The design session's brief to the bench work, 2026-11-14.
