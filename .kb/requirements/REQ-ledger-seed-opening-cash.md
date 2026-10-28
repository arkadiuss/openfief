---
title: Opening cash must be booked as a deposit, or a withdrawal for a negative balance, against the equity ledger account of its currency, keyed with the opening cash prefix so that the ledger can tell it from other deposits.
status: open
text_ref: docs/DESIGN.md#D-33
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-33
  - lane:strict
semantic_text: Opening cash must be booked as a deposit, or a withdrawal for a negative balance, against the equity ledger account of its currency, keyed with the opening cash prefix so that the ledger can tell it from other deposits.
logic_claims:
  - CLAIM-00CCC2DC91FA8392
semantic_clauses:
  - Opening cash must be booked as a deposit, or a withdrawal for a negative balance, against the equity ledger account of its currency, keyed with the opening cash prefix so that the ledger can tell it from other deposits.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 817cf5a9022da1993e005efe05506285511b32f6d7a928ba53301f1a5f4aff67
semantic_inventory:
  - claim_key: CLAIM-00CCC2DC91FA8392
    claim_text: Opening cash must be booked as a deposit, or a withdrawal for a negative balance, against the equity ledger account of its currency, keyed with the opening cash prefix so that the ledger can tell it from other deposits
    role: normative
    status: modeled
    span:
      start: 0
      end: 218
origin:
  kind: agent
  recorded_at: '2026-10-28T08:39:08.139Z'
id: REQ-ledger-seed-opening-cash
type: req
---
Opening cash must be booked as a deposit, or a withdrawal for a negative balance, against the equity ledger account of its currency, keyed with the opening cash prefix so that the ledger can tell it from other deposits.

## Context

D-23 seeds accounts with opening lots and the opening cash; D-33 makes the equity ledger account the counterpart of deposits, withdrawals and opening lots. The author told the orchestrator, in answer to this task (not yet written into DESIGN.md): opening cash stays a deposit or withdrawal against the equity ledger account, and seeding an account that is already seeded is refused with a clear error. The key prefix is how IsSeeding finds opening cash without a new transaction kind. Refines REQ-AUTO-3C2A493C661181C7. Modeled by Claude Code for the M0 seed-guard task; the author has not reviewed it yet.

## Source

> Besides D-19's accounts, each account has an equity ledger account per currency: the counterpart for deposits, withdrawals and opening lots.

docs/DESIGN.md D-33
