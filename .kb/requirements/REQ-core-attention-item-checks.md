---
title: Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks must all raise attention items.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-130-2
  - lane:strict
text_ref: design-decisions:D-130#2
origin:
  kind: import
  ref: design-decisions:D-130#2
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks must all raise attention items.
logic_claims:
  - CLAIM-26F4255156DB8E48
semantic_clauses:
  - Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks must all raise attention items.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: f5bda5098314f5c1be2d4b91ea1f9ca9ec76cb5a98870c62ecb98a27612f58a0
semantic_inventory:
  - claim_key: CLAIM-26F4255156DB8E48
    claim_text: Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks must all raise attention items
    role: normative
    status: modeled
    span:
      start: 0
      end: 217
id: REQ-core-attention-item-checks
type: req
---
Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks must all raise attention items.

## Context

D-130 puts every check that needs the owner into one append-only store of attention items (REQ-core-attention-item-store). The checks it names already exist, each with its own record: a flagged balance assertion (D-18, REQ-ledger-app-assertion-flag), a flagged price check (D-110, D-112), an ingest problem (REQ-ingest-problems-not-errors), an unmapped broker account (REQ-ingest-unmapped-broker-account), an unconfirmed XTB ticker (REQ-ingest-xtb-unconfirmed-ticker-waits), the owner's marginal PIT rate left unset (REQ-pl-ikze-liquidation-missing) and a confirmed transfer broken by a correction (REQ-ledger-transfer-broken-attention-item). The author asked for one common mechanism for every flag and failed check, which later agents can act on, while answering the implementation's questions on 2026-10-31, and recorded it as D-130.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. Flagged balance assertions and price checks, import problems, unmapped broker accounts, unconfirmed tickers, missing settings such as the PIT rate, broken transfer links and later checks all use it, and the UI lists them in one place. Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-130#2
