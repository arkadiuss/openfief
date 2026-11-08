---
title: After a bond is re-keyed to its registry ISIN, the Ghostfolio view's next sync must import and delete nothing, since a bond's asset is derived from its series code and its activities name transaction keys.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: After a bond is re-keyed to its registry ISIN, the Ghostfolio view's next sync must import and delete nothing, since a bond's asset is derived from its series code and its activities name transaction keys.
logic_claims:
  - CLAIM-8B086CCEFA5B7584
semantic_clauses:
  - After a bond is re-keyed to its registry ISIN, the Ghostfolio view's next sync must import and delete nothing, since a bond's asset is derived from its series code and its activities name transaction keys.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 851f29abfc6e8ffe905f521287dc820bc3157103c3a64abebfabaf9b156b914d
semantic_inventory:
  - claim_key: CLAIM-8B086CCEFA5B7584
    claim_text: After a bond is re-keyed to its registry ISIN, the Ghostfolio view's next sync must import and delete nothing, since a bond's asset is derived from its series code and its activities name transaction keys
    role: normative
    status: modeled
    span:
      start: 0
      end: 204
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:33.096Z'
id: REQ-views-ghostfolio-rekey-resync
type: req
---
After a bond is re-keyed to its registry ISIN, the Ghostfolio view's next sync must import and delete nothing, since a bond's asset is derived from its series code and its activities name transaction keys.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. D-64 derives a bond's MANUAL asset from the plugin instance and the series code. Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97), never one derived from its series code; the series code is its ticker.

docs/DESIGN.md#D-214
