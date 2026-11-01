---
title: The XTB connector must refuse an opening batch, because an xStation export holds neither the account's cash balance nor its positions' cost in the account currency.
status: closed
text_ref: docs/research.md#R-24
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-research-md-r-24
  - lane:strict
semantic_text: The XTB connector must refuse an opening batch, because an xStation export holds neither the account's cash balance nor its positions' cost in the account currency.
logic_claims:
  - CLAIM-E62C44890A0EF15D
semantic_clauses:
  - The XTB connector must refuse an opening batch, because an xStation export holds neither the account's cash balance nor its positions' cost in the account currency.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cb90f684fe78abd537953b19ce49f0c0ac86ec08b886acdbc51ceb5a9adf07cd
semantic_inventory:
  - claim_key: CLAIM-E62C44890A0EF15D
    claim_text: The XTB connector must refuse an opening batch, because an xStation export holds neither the account's cash balance nor its positions' cost in the account currency
    role: normative
    status: modeled
    span:
      start: 0
      end: 163
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:01.236Z'
id: REQ-ingest-xtb-opening-refused
type: req
---
The XTB connector must refuse an opening batch, because an xStation export holds neither the account's cash balance nor its positions' cost in the account currency.

## Context

D-23 seeds an account from lot-level positions and opening cash; R-24's export has no cash balance and no conversion rate in Open Positions, so the narrowest reading refuses opening mode (implementation questions). Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB's xStation "Historia operacji" export ... holds one account. Its labels are English, and it has three sheets, Closed Positions, Cash Operations and Open Positions.

docs/research.md#R-24
