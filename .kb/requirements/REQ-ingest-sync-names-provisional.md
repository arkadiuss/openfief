---
title: Every sync request to a connector bound to accounts must name each instrument those accounts post to under a provisional ISIN, held or not, once, and the treasury bond connector must answer the registry ISIN of each whose series the ministry's data lists.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: Every sync request to a connector bound to accounts must name each instrument those accounts post to under a provisional ISIN, held or not, once, and the treasury bond connector must answer the registry ISIN of each whose series the ministry's data lists.
logic_claims:
  - CLAIM-B3E620DB7CD33432
semantic_clauses:
  - Every sync request to a connector bound to accounts must name each instrument those accounts post to under a provisional ISIN, held or not, once, and the treasury bond connector must answer the registry ISIN of each whose series the ministry's data lists.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7b0200cdefb66b6092d241f99b4a9095a787c00234e1ea27e052447602ff90a9
semantic_inventory:
  - claim_key: CLAIM-B3E620DB7CD33432
    claim_text: Every sync request to a connector bound to accounts must name each instrument those accounts post to under a provisional ISIN, held or not, once, and the treasury bond connector must answer the registry ISIN of each whose series the ministry's data lists
    role: normative
    status: modeled
    span:
      start: 0
      end: 254
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:12.869Z'
id: REQ-ingest-sync-names-provisional
type: req
---
Every sync request to a connector bound to accounts must name each instrument those accounts post to under a provisional ISIN, held or not, once, and the treasury bond connector must answer the registry ISIN of each whose series the ministry's data lists.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A lot of a series the data doesn't list yet waits, like its terms, until its ISIN arrives, and a held series still without its ISIN or terms after seven days raises an attention item.

docs/DESIGN.md#D-214
