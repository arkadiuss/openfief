---
title: A lot of a series the Ministry of Finance's data doesn't list yet must wait, like its terms, until its ISIN arrives.
status: open
text_ref: design-decisions:D-214#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-214-3
  - lane:strict
semantic_text: A lot of a series the Ministry of Finance's data doesn't list yet must wait, like its terms, until its ISIN arrives.
logic_claims:
  - CLAIM-967C59B5531776CE
semantic_clauses:
  - A lot of a series the Ministry of Finance's data doesn't list yet must wait, like its terms, until its ISIN arrives.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bb0324c0e5044e6039fc94ef9335cae546b6c44adc029749fba3a2f1dad62ce0
semantic_inventory:
  - claim_key: CLAIM-967C59B5531776CE
    claim_text: A lot of a series the Ministry of Finance's data doesn't list yet must wait, like its terms, until its ISIN arrives
    role: normative
    status: modeled
    span:
      start: 0
      end: 115
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-214#3
  recorded_at: '2026-11-08T00:36:40.657Z'
id: REQ-treasurybonds-lot-waits-for-isin
type: req
---
A lot of a series the Ministry of Finance's data doesn't list yet must wait, like its terms, until its ISIN arrives.

## Context

D-99 had OpenFief derive a treasury bond's ISIN from its series code, because owners never see the registry's own ISIN. The Ministry of Finance's published retail bond data (D-97) lists every series' registry ISIN in its "Kod ISIN" column, and Ghostfolio, market data and OKI classification need the real one, so the author decided D-214 after the design session that followed the D-97 follow-ups: the registry's ISIN identifies a series, its series code becomes its ticker, and a lot of a series the data doesn't list yet waits for its ISIN as D-97's pending lot waits for its terms, with an attention item after seven days.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97), never one derived from its series code; the series code is its ticker. A lot of a series the data doesn't list yet waits, like its terms, until its ISIN arrives, and a held series still without its ISIN or terms after seven days raises an attention item.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-214#3
