---
title: 'The built-in UI must show a neutral pending pill, whose tooltip says it moves to the registry''s ISIN when the Ministry of Finance''s data lists the series, next to every instrument the read API marks isin_provisional wherever its ISIN shows: an account''s holdings and open lots, the holdings behind an allocation bar and the instrument page.'
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: 'The built-in UI must show a neutral pending pill, whose tooltip says it moves to the registry''s ISIN when the Ministry of Finance''s data lists the series, next to every instrument the read API marks isin_provisional wherever its ISIN shows: an account''s holdings and open lots, the holdings behind an allocation bar and the instrument page.'
logic_claims:
  - CLAIM-00F27C26C9B0035B
semantic_clauses:
  - 'The built-in UI must show a neutral pending pill, whose tooltip says it moves to the registry''s ISIN when the Ministry of Finance''s data lists the series, next to every instrument the read API marks isin_provisional wherever its ISIN shows: an account''s holdings and open lots, the holdings behind an allocation bar and the instrument page.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c9e33c5d3d42b5c431592fbad2c08db67088c3582972f625b72112d210bf6694
semantic_inventory:
  - claim_key: CLAIM-00F27C26C9B0035B
    claim_text: 'The built-in UI must show a neutral pending pill, whose tooltip says it moves to the registry''s ISIN when the Ministry of Finance''s data lists the series, next to every instrument the read API marks isin_provisional wherever its ISIN shows: an account''s holdings and open lots, the holdings behind an allocation bar and the instrument page'
    role: normative
    status: modeled
    span:
      start: 0
      end: 339
origin:
  kind: agent
  recorded_at: '2026-11-08T02:00:51.709Z'
id: REQ-ui-isin-pending-pill
type: req
---
The built-in UI must show a neutral pending pill, whose tooltip says it moves to the registry's ISIN when the Ministry of Finance's data lists the series, next to every instrument the read API marks isin_provisional wherever its ISIN shows: an account's holdings and open lots, the holdings behind an allocation bar and the instrument page.

## Context

D-214 books a lot of a series the Ministry of Finance's data doesn't list yet under a provisional ISIN until the registry's arrives (REQ-treasurybonds-lot-provisional-isin-until-listed). The design session asked, in the D-214 follow-ups, that the UI show a Pending pill next to such a bond wherever its ISIN appears, with a tooltip saying it moves to the registry's ISIN when the ministry's data lists it, using the design system's existing pill. Modeled by Claude Code for the D-214 follow-ups; the author has not reviewed the wording.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97); the series code is its ticker. A lot of a series the data doesn't list yet is booked under a provisional ISIN, `PL`, the series code, `00` and a check digit, and moves to the registry's ISIN when the data lists it.

docs/DESIGN.md#D-214
