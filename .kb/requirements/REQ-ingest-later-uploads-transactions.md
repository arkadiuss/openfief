---
title: Uploads after an account's opening upload must be imported as transactions.
status: open
text_ref: design-decisions:D-160#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-160-3
  - lane:strict
semantic_text: Uploads after an account's opening upload must be imported as transactions.
logic_claims:
  - CLAIM-959E5980A52B3221
semantic_clauses:
  - Uploads after an account's opening upload must be imported as transactions.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 162840a181394e9e1fd8b980dfcb7151697621facc8b9b05e37635f79cbea9d0
semantic_inventory:
  - claim_key: CLAIM-959E5980A52B3221
    claim_text: Uploads after an account's opening upload must be imported as transactions
    role: normative
    status: modeled
    span:
      start: 0
      end: 74
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-160#3
  recorded_at: '2026-11-02T00:54:48.339Z'
id: REQ-ingest-later-uploads-transactions
type: req
---
Uploads after an account's opening upload must be imported as transactions.

## Why this changed

On 2026-11-01 the author recorded D-160. REQ-ingest-opening-upload-starts-account, a default requirement the implementation agent recorded while building D-160 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-160 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. The core chooses an upload's mode, and the owner may override it (D-109). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An opening upload starts its account at once: the account's go-live is recorded with the upload, even when nothing can be booked yet. Positions waiting for their ticker's ISIN (D-116) join the opening when the owner confirms it, and later uploads are imported as transactions.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-160#3
