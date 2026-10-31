---
title: Each change to an attention item must also be published as an event from the core's outbox, so notifications and, in later milestones, agents can act on it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-130-4
  - lane:strict
text_ref: design-decisions:D-130#4
origin:
  kind: import
  ref: design-decisions:D-130#4
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Each change to an attention item must also be published as an event from the core's outbox, so notifications and, in later milestones, agents can act on it.
logic_claims:
  - CLAIM-C07F4925353EC67D
semantic_clauses:
  - Each change to an attention item must also be published as an event from the core's outbox, so notifications and, in later milestones, agents can act on it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e9fb5fb8631b5d662a344d647774426056af848c7d2a8029f0e0942162482797
semantic_inventory:
  - claim_key: CLAIM-C07F4925353EC67D
    claim_text: Each change to an attention item must also be published as an event from the core's outbox, so notifications and, in later milestones, agents can act on it
    role: normative
    status: modeled
    span:
      start: 0
      end: 155
id: REQ-core-attention-item-outbox-events
type: req
---
Each change to an attention item must also be published as an event from the core's outbox, so notifications and, in later milestones, agents can act on it.

## Context

D-130 keeps every check that needs the owner as an attention item in one append-only store (REQ-core-attention-item-store). On 2026-10-31, while answering the implementation's questions on docs/implementation-questions.md, the author asked for one common mechanism for every flag and failed check that later agents can act on, so each change to an item is also published as an event from the core's outbox: notifications can follow it now, and agents in later milestones. How agents would act on attention items is still a proposal (P-57), not a decision.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks all use it, and the UI lists them in one place. Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-130#4
