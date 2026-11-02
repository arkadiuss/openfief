---
title: Positions of an opening upload waiting for their ticker's ISIN must join the opening when the owner confirms it.
status: open
text_ref: design-decisions:D-160#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-160-2
  - lane:strict
semantic_text: Positions of an opening upload waiting for their ticker's ISIN must join the opening when the owner confirms it.
logic_claims:
  - CLAIM-21238A8B511C7DD6
semantic_clauses:
  - Positions of an opening upload waiting for their ticker's ISIN must join the opening when the owner confirms it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d9f88c0890341838768202c973e1a6a9ce90c6ba15f04e0eae8198aacf5df1dc
semantic_inventory:
  - claim_key: CLAIM-21238A8B511C7DD6
    claim_text: Positions of an opening upload waiting for their ticker's ISIN must join the opening when the owner confirms it
    role: normative
    status: modeled
    span:
      start: 0
      end: 111
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-160#2
  recorded_at: '2026-11-02T00:54:42.011Z'
id: REQ-ingest-waiting-positions-join-opening
type: req
---
Positions of an opening upload waiting for their ticker's ISIN must join the opening when the owner confirms it.

## Why this changed

On 2026-11-01 the author recorded D-160. REQ-ingest-opening-upload-starts-account, a default requirement the implementation agent recorded while building D-160 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-160 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. The owner confirms a ticker's ISIN once (D-116), and its entries wait until then. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An opening upload starts its account at once: the account's go-live is recorded with the upload, even when nothing can be booked yet. Positions waiting for their ticker's ISIN (D-116) join the opening when the owner confirms it, and later uploads are imported as transactions.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-160#2
