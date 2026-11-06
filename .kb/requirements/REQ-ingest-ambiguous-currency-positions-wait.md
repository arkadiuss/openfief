---
title: Positions of a ticker whose trading currency the owner must pick must wait until the owner picks it.
status: open
text_ref: design-decisions:D-170#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-170-2
  - lane:strict
semantic_text: Positions of a ticker whose trading currency the owner must pick must wait until the owner picks it.
logic_claims:
  - CLAIM-8600A64159C9C6C2
semantic_clauses:
  - Positions of a ticker whose trading currency the owner must pick must wait until the owner picks it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2054808cfb5bfde2245bda0daf57085327d1078135f0e11a9ac173b3f104371e
semantic_inventory:
  - claim_key: CLAIM-8600A64159C9C6C2
    claim_text: Positions of a ticker whose trading currency the owner must pick must wait until the owner picks it
    role: normative
    status: modeled
    span:
      start: 0
      end: 99
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-170#2
  recorded_at: '2026-11-02T02:36:00.282Z'
id: REQ-ingest-ambiguous-currency-positions-wait
type: req
---
Positions of a ticker whose trading currency the owner must pick must wait until the owner picks it.

## Why this changed

On 2026-11-01 the author recorded D-170, which narrows D-159 (the decision gives the reason: an exchange such as London quotes in pence, pounds and dollars). The implementation agent of D-170 (impl/d170) recorded default requirements for it before a firm one existed: REQ-ingest-ticker-currency-picked, REQ-ingest-ticker-currency-pick-required and REQ-ui-ticker-currency-choice. The firm requirements for D-170 replace them, and they are closed; their scenarios and the code that implements them move here, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. Rows of an unconfirmed ticker already wait as import problems (D-116). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> When the listing gives no currency and the exchange its market suffix names quotes in more than one, as London does in pence, pounds and dollars, the owner picks the trading currency when confirming the ticker's ISIN (D-116), and its positions wait until then; no currency is assumed.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-170#2
