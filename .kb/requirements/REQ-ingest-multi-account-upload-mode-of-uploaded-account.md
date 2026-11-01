---
title: An upload of a statement covering several accounts must run in the mode chosen for the account it was uploaded for, which the owner can override.
status: open
text_ref: design-decisions:D-140#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-140-1
  - lane:strict
semantic_text: An upload of a statement covering several accounts must run in the mode chosen for the account it was uploaded for, which the owner can override.
logic_claims:
  - CLAIM-FF05AFDD4B3E5FE8
semantic_clauses:
  - An upload of a statement covering several accounts must run in the mode chosen for the account it was uploaded for, which the owner can override.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 720a83add0f772fbb606f8a0d25789d962130005627ecaec46aeab600c7e3cd9
semantic_inventory:
  - claim_key: CLAIM-FF05AFDD4B3E5FE8
    claim_text: An upload of a statement covering several accounts must run in the mode chosen for the account it was uploaded for, which the owner can override
    role: normative
    status: modeled
    span:
      start: 0
      end: 144
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-140#1
  recorded_at: '2026-11-01T12:23:47.526Z'
id: REQ-ingest-multi-account-upload-mode-of-uploaded-account
type: req
---
An upload of a statement covering several accounts must run in the mode chosen for the account it was uploaded for, which the owner can override.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #24. D-109 has the core choose an upload's mode, opening for an account not yet seeded and transactions after that, with an override for the owner (REQ-ingest-statement-mode-chosen-by-core). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, which the owner can override: an opening upload seeds only the accounts not seeded yet, and a transactions upload books every mapped account.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-140#1
