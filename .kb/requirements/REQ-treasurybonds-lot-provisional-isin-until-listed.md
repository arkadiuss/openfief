---
title: A lot of a series the Ministry of Finance's data doesn't list yet must be booked under a provisional ISIN, PL, the series code, 00 and a check digit, and must move to the registry's ISIN when the data lists it.
status: open
text_ref: design-decisions:D-214#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-214-3
  - lane:strict
semantic_text: A lot of a series the Ministry of Finance's data doesn't list yet must be booked under a provisional ISIN, PL, the series code, 00 and a check digit, and must move to the registry's ISIN when the data lists it.
logic_claims:
  - CLAIM-1331A4B38FD77167
semantic_clauses:
  - A lot of a series the Ministry of Finance's data doesn't list yet must be booked under a provisional ISIN, PL, the series code, 00 and a check digit, and must move to the registry's ISIN when the data lists it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 05278f07d1798cf60fd9cbe8ae3f5fe57a96c0e738b926c6fffb7a97bb660c78
semantic_inventory:
  - claim_key: CLAIM-1331A4B38FD77167
    claim_text: A lot of a series the Ministry of Finance's data doesn't list yet must be booked under a provisional ISIN, PL, the series code, 00 and a check digit, and must move to the registry's ISIN when the data lists it
    role: normative
    status: modeled
    span:
      start: 0
      end: 209
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-214#3
  recorded_at: '2026-11-08T01:31:54.370Z'
id: REQ-treasurybonds-lot-provisional-isin-until-listed
type: req
---
A lot of a series the Ministry of Finance's data doesn't list yet must be booked under a provisional ISIN, PL, the series code, 00 and a check digit, and must move to the registry's ISIN when the data lists it.

## Context

D-99 had OpenFief derive a treasury bond's ISIN from its series code, because owners never see the registry's own ISIN. The Ministry of Finance's published retail bond data (D-97) lists every series' registry ISIN in its "Kod ISIN" column, and Ghostfolio, market data and OKI classification need the real one, so the author decided D-214: the registry's ISIN identifies a series and its series code becomes its ticker. The author then reworded D-214 to say what was built, a lot of a series the data doesn't list yet booked under a provisional ISIN that moves to the registry's when the data lists it, and deleted D-99.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97); the series code is its ticker. A lot of a series the data doesn't list yet is booked under a provisional ISIN, `PL`, the series code, `00` and a check digit, and moves to the registry's ISIN when the data lists it. A held series still without its ISIN or terms after seven days raises an attention item.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-214#3

## Why this changed

REQ-treasurybonds-lot-waits-for-isin was imported from D-214's first wording, where a lot of a series the data doesn't list yet waits, like its terms, until its ISIN arrives. What was built books the lot at once under a provisional ISIN, PL, the series code, 00 and a check digit, and re-keys it to the registry's ISIN when a sync brings it (REQ-treasurybonds-provisional-isin, REQ-ingest-rekey-provisional-instrument). The author reworded D-214 to say so, replacing "waits", so the requirement is imported again from the new text.
