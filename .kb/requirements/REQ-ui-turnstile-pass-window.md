---
title: 'A visitor who passed Turnstile is not checked again for 30 minutes after their last model question and at most 2 hours from the check: each model question that reaches the model renews the 30 minutes, the renewal never reaches past 2 hours from the original check, and a pass past either limit is refused, so the question is checked on demand again.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'A visitor who passed Turnstile is not checked again for 30 minutes after their last model question and at most 2 hours from the check: each model question that reaches the model renews the 30 minutes, the renewal never reaches past 2 hours from the original check, and a pass past either limit is refused, so the question is checked on demand again.'
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
tags:
  - turnstile
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 01450566714a9b2b87331efa8e4d0fd0ae852fa00a45cd9eb14fb54d32ff305b
logic_claims:
  - CLAIM-3F4662555448E36E
semantic_inventory:
  - claim_key: CLAIM-3F4662555448E36E
    claim_text: 'A visitor who passed Turnstile is not checked again for 30 minutes after their last model question and at most 2 hours from the check: each model question that reaches the model renews the 30 minutes, the renewal never reaches past 2 hours from the original check, and a pass past either limit is refused, so the question is checked on demand again'
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 348
    subjects:
      - turnstile.pass_window
    reason: It states how the built-in UI behaves; no predicate of the ontology states it yet.
rationale: D-208, as the design session amended it, spares a visitor who passed Turnstile a new check on every model question, since the Ask box made one for each. The window slides with the questions up to a cap, so a pass is not a session-long licence. The windows are D-208's own, 30 minutes and 2 hours.
id: REQ-ui-turnstile-pass-window
type: req
---
A visitor who passed Turnstile is not checked again for 30 minutes after their last model question and at most 2 hours from the check: each model question that reaches the model renews the 30 minutes, the renewal never reaches past 2 hours from the original check, and a pass past either limit is refused, so the question is checked on demand again.

## Context

D-208, as the design session amended it, spares a visitor who passed Turnstile a new check on every model question, since the Ask box made one for each. The window slides with the questions up to a cap, so a pass is not a session-long licence. The windows are D-208's own, 30 minutes and 2 hours.

## Source

> D-208 (docs/DESIGN.md), as amended: A visitor who passes Turnstile isn't checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question.
