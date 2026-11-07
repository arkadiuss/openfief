---
title: A re-key must move a bond series' stored terms, exemption class, prices, cross-check prices, price check threshold, prices source and model portfolio target to its registry ISIN, the registry's own rows winning and a target's weight joining, and change nothing when run again.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: A re-key must move a bond series' stored terms, exemption class, prices, cross-check prices, price check threshold, prices source and model portfolio target to its registry ISIN, the registry's own rows winning and a target's weight joining, and change nothing when run again.
logic_claims:
  - CLAIM-A8864A732F7E1EF6
semantic_clauses:
  - A re-key must move a bond series' stored terms, exemption class, prices, cross-check prices, price check threshold, prices source and model portfolio target to its registry ISIN, the registry's own rows winning and a target's weight joining, and change nothing when run again.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 16aafba19f7c29147f6290f75ec6954a1ef4907ba32f4c5d1d615cc8a8f4775f
semantic_inventory:
  - claim_key: CLAIM-A8864A732F7E1EF6
    claim_text: A re-key must move a bond series' stored terms, exemption class, prices, cross-check prices, price check threshold, prices source and model portfolio target to its registry ISIN, the registry's own rows winning and a target's weight joining, and change nothing when run again
    role: normative
    status: modeled
    span:
      start: 0
      end: 275
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:27.787Z'
id: REQ-valuation-rekey-inputs
type: req
---
A re-key must move a bond series' stored terms, exemption class, prices, cross-check prices, price check threshold, prices source and model portfolio target to its registry ISIN, the registry's own rows winning and a target's weight joining, and change nothing when run again.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97), never one derived from its series code; the series code is its ticker.

docs/DESIGN.md#D-214
