---
title: An import of the issuer's own export must derive each bond's ISIN from its series code the same way, so one series is never held twice.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-99-2
  - lane:strict
text_ref: design-decisions:D-99#2
origin:
  kind: import
  ref: design-decisions:D-99#2
  recorded_at: '2026-10-29T09:50:06.888Z'
semantic_text: An import of the issuer's own export must derive each bond's ISIN from its series code the same way, so one series is never held twice.
logic_claims:
  - CLAIM-B03098C123D58F53
semantic_clauses:
  - An import of the issuer's own export must derive each bond's ISIN from its series code the same way, so one series is never held twice.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0efd1589846c89dd1be66e714f076d70ce18d9c431971f823b4156a089c1ddc0
semantic_inventory:
  - claim_key: CLAIM-B03098C123D58F53
    claim_text: An import of the issuer's own export must derive each bond's ISIN from its series code the same way, so one series is never held twice
    role: normative
    status: modeled
    span:
      start: 0
      end: 134
id: REQ-ingest-bond-export-same-isin
type: req
---
An import of the issuer's own export must derive each bond's ISIN from its series code the same way, so one series is never held twice.

## Context

D-9 adds an import of the bond service's own export if one exists (OQ-22), next to manual lot entry. D-99 derives a treasury bond's ISIN from its series code (REQ-pl-bond-isin-from-series) and has that import use the same derivation, so a series entered by hand and the same series in the export are one instrument, never two holdings. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> Owners never see the registry's own ISIN for a treasury bond, so OpenFief derives one from the series code: `PL`, the series code, `00` and a check digit, such as `PLEDO0536000` for EDO0536. An import of the issuer's own export uses the same derivation, so one series is never held twice.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-99#2
