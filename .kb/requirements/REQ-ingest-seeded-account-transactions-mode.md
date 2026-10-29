---
title: The core must sync an account already seeded from opening lots in transactions mode on its own, with no manual step switching it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-95-1
  - lane:strict
text_ref: design-decisions:D-95#1
origin:
  kind: import
  ref: design-decisions:D-95#1
  recorded_at: '2026-10-29T09:40:18.797Z'
semantic_text: The core must sync an account already seeded from opening lots in transactions mode on its own, with no manual step switching it.
logic_claims:
  - CLAIM-E1C536B7F5B541E2
semantic_clauses:
  - The core must sync an account already seeded from opening lots in transactions mode on its own, with no manual step switching it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 55a692859fdf08c9aab0c17fa32df9bc2311f655fb70c3e36ce1ba31974e64b5
semantic_inventory:
  - claim_key: CLAIM-E1C536B7F5B541E2
    claim_text: The core must sync an account already seeded from opening lots in transactions mode on its own, with no manual step switching it
    role: normative
    status: modeled
    span:
      start: 0
      end: 128
id: REQ-ingest-seeded-account-transactions-mode
type: req
---
The core must sync an account already seeded from opening lots in transactions mode on its own, with no manual step switching it.

## Context

D-23 seeds accounts from each broker's current lot-level positions and takes transactions from go-live (REQ-AUTO-3C2A493C661181C7), and D-49 has the core own connector schedules. D-95 has the core move a seeded account to transactions mode by itself, so the owner never switches modes by hand. REQ-ingest-scheduler-due, which the author has not reviewed, decides the mode per connector instance, opening until an opening batch completes, rather than by whether the account is already seeded. The author recorded D-93 to D-95 after the design session that answered the implementation's questions.

## Source

> When an account has already been seeded from opening lots, the core syncs it in transactions mode on its own; no manual step switches it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-95#1
