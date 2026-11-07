---
title: A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data, never by one derived from its series code.
status: open
text_ref: design-decisions:D-214#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-214-1
  - lane:strict
semantic_text: A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data, never by one derived from its series code.
logic_claims:
  - CLAIM-7CEF3D064E3E388C
semantic_clauses:
  - A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data, never by one derived from its series code.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4a912acad5daef86d4ae59d92abed29dd13f1cf2177637f5689e8051bf90499e
semantic_inventory:
  - claim_key: CLAIM-7CEF3D064E3E388C
    claim_text: A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data, never by one derived from its series code
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-214#1
  recorded_at: '2026-11-07T22:44:47.513Z'
id: REQ-treasurybonds-registry-isin
type: req
---
A treasury bond must be identified by its real ISIN, the registry's, from the Ministry of Finance's data, never by one derived from its series code.

## Context

D-99 had OpenFief derive a treasury bond's ISIN from its series code, because owners never see the registry's own ISIN. The Ministry of Finance's published retail bond data (D-97) lists every series' registry ISIN in its "Kod ISIN" column, and Ghostfolio, market data and OKI classification need the real one, so the author decided D-214 after the design session that followed the D-97 follow-ups: the registry's ISIN identifies a series, its series code becomes its ticker, and a lot of a series the data doesn't list yet waits for its ISIN as D-97's pending lot waits for its terms, with an attention item after seven days.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97), never one derived from its series code; the series code is its ticker. A lot of a series the data doesn't list yet waits, like its terms, until its ISIN arrives, and a held series still without its ISIN or terms after seven days raises an attention item.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-214#1

## Why this changed

REQ-pl-bond-isin-from-series and REQ-ingest-bond-export-same-isin were imported from D-99, which had OpenFief derive every treasury bond's ISIN from its series code, PL, the code, 00 and a check digit, and an import of the issuer's own export derive it the same way. D-214 replaces that: a series is identified by the registry's ISIN from the Ministry of Finance's data, never a derived one, so an import of the issuer's export names the same series by the same registry ISIN. The derived form survives only as the provisional ISIN a series is held under until its registry ISIN arrives, which the core then re-keys (REQ-treasurybonds-provisional-isin).
