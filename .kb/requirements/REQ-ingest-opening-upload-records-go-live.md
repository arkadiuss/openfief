---
title: An opening upload must start its account at once, recording the account's go-live with the upload even when nothing can be booked yet.
status: open
text_ref: design-decisions:D-160#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-160-1
  - lane:strict
semantic_text: An opening upload must start its account at once, recording the account's go-live with the upload even when nothing can be booked yet.
logic_claims:
  - CLAIM-5A54B221B0228C09
semantic_clauses:
  - An opening upload must start its account at once, recording the account's go-live with the upload even when nothing can be booked yet.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 53013b8258c136872cfad5075859c1e4653b1a06fb3f29265cc61d8d7a1c8e95
semantic_inventory:
  - claim_key: CLAIM-5A54B221B0228C09
    claim_text: An opening upload must start its account at once, recording the account's go-live with the upload even when nothing can be booked yet
    role: normative
    status: modeled
    span:
      start: 0
      end: 133
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-160#1
  recorded_at: '2026-11-02T00:54:36.089Z'
id: REQ-ingest-opening-upload-records-go-live
type: req
---
An opening upload must start its account at once, recording the account's go-live with the upload even when nothing can be booked yet.

## Why this changed

On 2026-11-01 the author recorded D-160. REQ-ingest-opening-upload-starts-account, a default requirement the implementation agent recorded while building D-160 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-160 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. The go-live is D-23's; transactions are recorded from it on. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An opening upload starts its account at once: the account's go-live is recorded with the upload, even when nothing can be booked yet. Positions waiting for their ticker's ISIN (D-116) join the opening when the owner confirms it, and later uploads are imported as transactions.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-160#1
