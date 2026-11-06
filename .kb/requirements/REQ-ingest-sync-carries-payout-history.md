---
title: Each sync request must carry each bound account's payout setting with the day it took effect and the settings in force before it, and a plugin must refuse settings out of order.
status: closed
text_ref: docs/DESIGN.md#D-162
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-162
  - lane:strict
semantic_text: Each sync request must carry each bound account's payout setting with the day it took effect and the settings in force before it, and a plugin must refuse settings out of order.
logic_claims:
  - CLAIM-E0C05AAE0D849132
semantic_clauses:
  - Each sync request must carry each bound account's payout setting with the day it took effect and the settings in force before it, and a plugin must refuse settings out of order.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: df9f30317b76d25d73c432c34b915b33e934f6874e6a8a3cc63158f7737962d0
semantic_inventory:
  - claim_key: CLAIM-E0C05AAE0D849132
    claim_text: Each sync request must carry each bound account's payout setting with the day it took effect and the settings in force before it, and a plugin must refuse settings out of order
    role: normative
    status: modeled
    span:
      start: 0
      end: 176
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:12.427Z'
id: REQ-ingest-sync-carries-payout-history
type: req
---
Each sync request must carry each bound account's payout setting with the day it took effect and the settings in force before it, and a plugin must refuse settings out of order.

## Context

D-162 makes what fell due before a change follow the setting in force then, so the connector needs every setting to book a coupon or maturity on any day. Recorded by the implementation agent of D-159 to D-162 (impl/d159-162) with default firmness: no firm requirement for these decisions existed in kibi when they were built, so this traces the code to the decision directly until the design session imports the firm ones.

## Why this changed

The author decided this in D-162, which replaces the narrower reading REQ-ingest-sync-carries-payout recorded before the decision existed, so that requirement is closed.

## Source

> An account's first payout setting covers its whole history. A change takes effect from a date the owner gives, today by default and possibly earlier, and the next sync books the coupons and maturities after that date anew, superseding the earlier ones (D-18, D-143).

docs/DESIGN.md#D-162
