---
title: 'OpenFief must derive a treasury bond''s ISIN from its series code: PL, the series code, 00 and a check digit, such as PLEDO0536000 for EDO0536.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-99-1
  - lane:strict
text_ref: design-decisions:D-99#1
origin:
  kind: import
  ref: design-decisions:D-99#1
  recorded_at: '2026-10-29T09:49:21.141Z'
semantic_text: 'OpenFief must derive a treasury bond''s ISIN from its series code: PL, the series code, 00 and a check digit, such as PLEDO0536000 for EDO0536.'
logic_claims:
  - CLAIM-72D32AB9D2DE37EC
semantic_clauses:
  - 'OpenFief must derive a treasury bond''s ISIN from its series code: PL, the series code, 00 and a check digit, such as PLEDO0536000 for EDO0536.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 3d36b1fd1d111523718e2ce0c61dff73e9510ecd965f97fd271dd96650531749
semantic_inventory:
  - claim_key: CLAIM-72D32AB9D2DE37EC
    claim_text: 'OpenFief must derive a treasury bond''s ISIN from its series code: PL, the series code, 00 and a check digit, such as PLEDO0536000 for EDO0536'
    role: normative
    status: modeled
    span:
      start: 0
      end: 141
id: REQ-pl-bond-isin-from-series
type: req
---
OpenFief must derive a treasury bond's ISIN from its series code: PL, the series code, 00 and a check digit, such as PLEDO0536000 for EDO0536.

## Context

D-35 identifies every instrument by an ISIN with a valid check digit plus a venue symbol, but owners never see the registry's own ISIN for a treasury bond; they know the series code, such as EDO0536. D-99 has OpenFief derive the ISIN from the series code instead, and REQ-ingest-bond-export-same-isin applies the same derivation to the issuer's own export. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> Owners never see the registry's own ISIN for a treasury bond, so OpenFief derives one from the series code: `PL`, the series code, `00` and a check digit, such as `PLEDO0536000` for EDO0536. An import of the issuer's own export uses the same derivation, so one series is never held twice.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-99#1
