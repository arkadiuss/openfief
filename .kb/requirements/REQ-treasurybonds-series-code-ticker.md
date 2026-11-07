---
title: A treasury bond's series code must be its ticker.
status: open
text_ref: design-decisions:D-214#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-214-2
  - lane:strict
semantic_text: A treasury bond's series code must be its ticker.
logic_claims:
  - CLAIM-D6795C323BFCBE03
semantic_clauses:
  - A treasury bond's series code must be its ticker.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7b8df4b430998a4cb5019eef17b1f638262fc111778e0516860c82207d19065a
semantic_inventory:
  - claim_key: CLAIM-D6795C323BFCBE03
    claim_text: A treasury bond's series code must be its ticker
    role: normative
    status: modeled
    span:
      start: 0
      end: 48
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-214#2
  recorded_at: '2026-11-07T22:44:51.786Z'
id: REQ-treasurybonds-series-code-ticker
type: req
---
A treasury bond's series code must be its ticker.

## Context

D-99 had OpenFief derive a treasury bond's ISIN from its series code, because owners never see the registry's own ISIN. The Ministry of Finance's published retail bond data (D-97) lists every series' registry ISIN in its "Kod ISIN" column, and Ghostfolio, market data and OKI classification need the real one, so the author decided D-214 after the design session that followed the D-97 follow-ups: the registry's ISIN identifies a series, its series code becomes its ticker, and a lot of a series the data doesn't list yet waits for its ISIN as D-97's pending lot waits for its terms, with an attention item after seven days.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97), never one derived from its series code; the series code is its ticker. A lot of a series the data doesn't list yet waits, like its terms, until its ISIN arrives, and a held series still without its ISIN or terms after seven days raises an attention item.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-214#2
