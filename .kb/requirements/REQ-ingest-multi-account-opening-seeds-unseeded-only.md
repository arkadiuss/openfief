---
title: An opening upload of a statement covering several accounts must seed only the accounts not seeded yet.
status: open
text_ref: design-decisions:D-140#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-140-2
  - lane:strict
semantic_text: An opening upload of a statement covering several accounts must seed only the accounts not seeded yet.
logic_claims:
  - CLAIM-98677A832F471470
semantic_clauses:
  - An opening upload of a statement covering several accounts must seed only the accounts not seeded yet.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f68b68d2d230fd552a9fdf2ffdd4f90c85919a05fda9d8d7b3ea3aeba33014f7
semantic_inventory:
  - claim_key: CLAIM-98677A832F471470
    claim_text: An opening upload of a statement covering several accounts must seed only the accounts not seeded yet
    role: normative
    status: modeled
    span:
      start: 0
      end: 101
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-140#2
  recorded_at: '2026-11-01T12:23:57.783Z'
id: REQ-ingest-multi-account-opening-seeds-unseeded-only
type: req
---
An opening upload of a statement covering several accounts must seed only the accounts not seeded yet.

## Why this changed

On 2026-11-01 the author recorded D-140, answering the implementation's question #24: an upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, so an opening upload seeds only the accounts not seeded yet and a transactions upload books every mapped account. Reason not stated. REQ-ingest-statement-multi-account had such a statement book every mapped account whatever the mode, so this requirement replaces it together with REQ-ingest-multi-account-transactions-books-every-mapped and REQ-ingest-statement-multi-account-reports-unmapped.

## Context

An IBKR Flex statement of linked accounts is such a statement (D-109). D-23 seeds an account from its opening lots, and an opening upload seeds the account for scheduled syncs too (REQ-ingest-opening-upload-seeds-syncs). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, which the owner can override: an opening upload seeds only the accounts not seeded yet, and a transactions upload books every mapped account.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-140#2
