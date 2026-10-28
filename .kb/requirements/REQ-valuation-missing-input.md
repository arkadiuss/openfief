---
title: 'A missing price, FX rate or index reading must not fail a revaluation: the holding valuation must record which input is missing and why, with no value, and the account''s valued-through date must be the day before its first missing input, or today when none is missing, keeping an earlier gap the revaluation did not reach.'
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: 'A missing price, FX rate or index reading must not fail a revaluation: the holding valuation must record which input is missing and why, with no value, and the account''s valued-through date must be the day before its first missing input, or today when none is missing, keeping an earlier gap the revaluation did not reach.'
logic_claims:
  - CLAIM-FBF72D109CB01BB4
semantic_clauses:
  - 'A missing price, FX rate or index reading must not fail a revaluation: the holding valuation must record which input is missing and why, with no value, and the account''s valued-through date must be the day before its first missing input, or today when none is missing, keeping an earlier gap the revaluation did not reach.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fe80dee1c229a8c290a8123ed77761eb086de4572875fdd58e12a6b2366f4a6e
semantic_inventory:
  - claim_key: CLAIM-FBF72D109CB01BB4
    claim_text: 'A missing price, FX rate or index reading must not fail a revaluation: the holding valuation must record which input is missing and why, with no value, and the account''s valued-through date must be the day before its first missing input, or today when none is missing, keeping an earlier gap the revaluation did not reach'
    role: normative
    status: modeled
    span:
      start: 0
      end: 321
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:32.439Z'
id: REQ-valuation-missing-input
type: req
---
A missing price, FX rate or index reading must not fail a revaluation: the holding valuation must record which input is missing and why, with no value, and the account's valued-through date must be the day before its first missing input, or today when none is missing, keeping an earlier gap the revaluation did not reach.

## Context

The read API reports the date each account is valued through (D-36); a gap in market data must show there rather than stop every later day from being valued. The rule for an earlier gap keeps a revaluation that starts after it from hiding it. Reason for the exact rule not stated by the author. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> The read API reports, per account, the date it is valued through and whether a revaluation is pending or running

docs/DESIGN.md D-36
