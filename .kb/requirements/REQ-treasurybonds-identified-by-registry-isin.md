---
title: A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data.
status: open
text_ref: design-decisions:D-214#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-214-1
  - lane:strict
semantic_text: A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data.
logic_claims:
  - CLAIM-F19256CC38B81B72
semantic_clauses:
  - A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d1071ee277f928d58f1802cd7f9bb4cda126102a1b83917ed5fbbea54e77df92
semantic_inventory:
  - claim_key: CLAIM-F19256CC38B81B72
    claim_text: A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data
    role: normative
    status: modeled
    span:
      start: 0
      end: 104
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-214#1
  recorded_at: '2026-11-08T01:31:49.453Z'
id: REQ-treasurybonds-identified-by-registry-isin
type: req
---
A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data.

## Context

D-99 had OpenFief derive a treasury bond's ISIN from its series code, because owners never see the registry's own ISIN. The Ministry of Finance's published retail bond data (D-97) lists every series' registry ISIN in its "Kod ISIN" column, and Ghostfolio, market data and OKI classification need the real one, so the author decided D-214: the registry's ISIN identifies a series and its series code becomes its ticker. The author then reworded D-214 to say what was built, a lot of a series the data doesn't list yet booked under a provisional ISIN that moves to the registry's when the data lists it, and deleted D-99.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97); the series code is its ticker. A lot of a series the data doesn't list yet is booked under a provisional ISIN, `PL`, the series code, `00` and a check digit, and moves to the registry's ISIN when the data lists it. A held series still without its ISIN or terms after seven days raises an attention item.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-214#1

## Why this changed

REQ-treasurybonds-registry-isin was imported from D-214's first wording, which added "never one derived from its series code". That contradicted the provisional booking D-214's third clause now states, where a lot of an unlisted series is held under an ISIN derived from its series code until the registry's arrives (REQ-treasurybonds-lot-provisional-isin-until-listed). The author's rewording drops the phrase, so the requirement is imported again from the new text.
