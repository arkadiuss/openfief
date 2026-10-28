---
title: A sleeve must have an ID, a name and exactly one role, defensive or offensive, and any other sleeve must be rejected.
status: open
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
semantic_text: A sleeve must have an ID, a name and exactly one role, defensive or offensive, and any other sleeve must be rejected.
logic_claims:
  - CLAIM-312615E409AAC37E
semantic_clauses:
  - A sleeve must have an ID, a name and exactly one role, defensive or offensive, and any other sleeve must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c252842dc54653dfc38a8a5f7e175186431d6eca626a1865e492c33a4f12912f
semantic_inventory:
  - claim_key: CLAIM-312615E409AAC37E
    claim_text: A sleeve must have an ID, a name and exactly one role, defensive or offensive, and any other sleeve must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
origin:
  kind: agent
  recorded_at: '2026-10-28T05:13:31.021Z'
id: REQ-portfolio-sleeve-role
type: req
---
A sleeve must have an ID, a name and exactly one role, defensive or offensive, and any other sleeve must be rejected.

## Context

The glossary defines a sleeve as a part of the portfolio with one role: defensive sleeves are never available to agents, and the offensive sleeve is what fiefs are leased from (GR-6). Accounts belong to exactly one sleeve, so allocation by sleeve (D-30) and the later lease ledger depend on every sleeve having a definite role. In M0 the portfolio context holds sleeves only. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Each thesis agent has a budget leased from the offensive part of the portfolio, with strong guarantees that the agent can never get more of the portfolio than its lease.

docs/DESIGN.md GR-6

> A part of the portfolio with one role. Defensive: not available to agents. Offensive: the part fiefs are leased from.

docs/glossary.md (Sleeve)
