---
title: Every check that needs the owner must raise an attention item in one append-only store that keeps its kind, subject, state (open, acknowledged or resolved) and history.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-130-1
  - lane:strict
text_ref: design-decisions:D-130#1
origin:
  kind: import
  ref: design-decisions:D-130#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Every check that needs the owner must raise an attention item in one append-only store that keeps its kind, subject, state (open, acknowledged or resolved) and history.
logic_claims:
  - CLAIM-C09B827CC1A61D1B
semantic_clauses:
  - Every check that needs the owner must raise an attention item in one append-only store that keeps its kind, subject, state (open, acknowledged or resolved) and history.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: ed7e398af7d40da043c115b5a0f7747af6edcc9b348b8dd7ab7c020b6e6d4293
semantic_inventory:
  - claim_key: CLAIM-C09B827CC1A61D1B
    claim_text: Every check that needs the owner must raise an attention item in one append-only store that keeps its kind, subject, state (open, acknowledged or resolved) and history
    role: normative
    status: modeled
    span:
      start: 0
      end: 167
id: REQ-core-attention-item-store
type: req
---
Every check that needs the owner must raise an attention item in one append-only store that keeps its kind, subject, state (open, acknowledged or resolved) and history.

## Context

Until D-130 each check that needed the owner kept its own record and page: a flagged balance assertion is stored as a check result (REQ-ledger-app-assertion-flag), a flagged price check has its own acknowledgements (REQ-valuation-price-check-acknowledge), import problems, unmapped broker accounts (REQ-ingest-unmapped-broker-account) and unconfirmed XTB tickers (REQ-ingest-xtb-unconfirmed-ticker-waits) each show elsewhere, and a missing PIT rate only leaves IKZE's liquidation value missing (REQ-pl-ikze-liquidation-missing). While answering the implementation's questions on docs/implementation-questions.md on 2026-10-31, among them #17, where one option was to flag a broken transfer link like a flagged balance assertion, the author asked for one common mechanism for every flag and failed check that later agents can act on, and recorded it as D-130. REQ-core-attention-item-checks lists the checks that use it.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks all use it, and the UI lists them in one place. Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-130#1
