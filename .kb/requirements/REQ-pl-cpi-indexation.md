---
title: From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin with a negative CPI counted as zero, a ROR or DOR bond's rate must be the NBP reference rate plus the margin, and OTS and TOS must keep their first-period rate.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin with a negative CPI counted as zero, a ROR or DOR bond's rate must be the NBP reference rate plus the margin, and OTS and TOS must keep their first-period rate.
logic_claims:
  - CLAIM-5C34FC18CE32ECEA
semantic_clauses:
  - From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin with a negative CPI counted as zero, a ROR or DOR bond's rate must be the NBP reference rate plus the margin, and OTS and TOS must keep their first-period rate.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d81dc7db7fb42792e36d9fbf6d186293460dacc615192b1cc72cd4d79cebfab8
semantic_inventory:
  - claim_key: CLAIM-5C34FC18CE32ECEA
    claim_text: From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin with a negative CPI counted as zero, a ROR or DOR bond's rate must be the NBP reference rate plus the margin, and OTS and TOS must keep their first-period rate
    role: normative
    status: modeled
    span:
      start: 0
      end: 272
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:54.508Z'
id: REQ-pl-cpi-indexation
type: req
---
From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin with a negative CPI counted as zero, a ROR or DOR bond's rate must be the NBP reference rate plus the margin, and OTS and TOS must keep their first-period rate.

## Context

The bonds connector takes CPI and the NBP reference rate as inputs (DESIGN.md 3 diagram: treasury bonds, CPI + NBP ref. rate). Readings and margins are data per series and per period, never hard-coded. The deflation floor and the CPI month (announced by GUS in the month before the period) come from the product descriptions. Refines REQ-AUTO-30FD669113CEFE3B. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet. obligacjeskarbowe.pl and gov.pl were not reachable from the agent's sandbox, so the rule was checked only against search-result excerpts of issue letters and secondary calculators.

## Source

> Connector: treasury bonds<br/>manual lots, CPI + NBP ref. rate

docs/DESIGN.md section 3 diagram

Product pages: https://www.obligacjeskarbowe.pl/oferta-obligacji/ and https://www.obligacjeskarbowe.pl/listy-emisyjne/
