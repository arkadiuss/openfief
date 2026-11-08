---
title: 'The end-to-end institution fixture must serve the treasury-bonds connector''s sources in their real formats: the Ministry of Finance''s gov.pl page and dane.gov.pl resource list leading to a synthetic .xls with each series'' registry ISIN, rates and margin, withheld until the suite publishes it, GUS''s CPI and NBP''s interest rate archive.'
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: 'The end-to-end institution fixture must serve the treasury-bonds connector''s sources in their real formats: the Ministry of Finance''s gov.pl page and dane.gov.pl resource list leading to a synthetic .xls with each series'' registry ISIN, rates and margin, withheld until the suite publishes it, GUS''s CPI and NBP''s interest rate archive.'
logic_claims:
  - CLAIM-96B8340BEE8F155C
semantic_clauses:
  - 'The end-to-end institution fixture must serve the treasury-bonds connector''s sources in their real formats: the Ministry of Finance''s gov.pl page and dane.gov.pl resource list leading to a synthetic .xls with each series'' registry ISIN, rates and margin, withheld until the suite publishes it, GUS''s CPI and NBP''s interest rate archive.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6497ee77b09bff68b820e96653163081ef3577e1b5f4de483d7fcbbbf76e642b
semantic_inventory:
  - claim_key: CLAIM-96B8340BEE8F155C
    claim_text: 'The end-to-end institution fixture must serve the treasury-bonds connector''s sources in their real formats: the Ministry of Finance''s gov.pl page and dane.gov.pl resource list leading to a synthetic .xls with each series'' registry ISIN, rates and margin, withheld until the suite publishes it, GUS''s CPI and NBP''s interest rate archive'
    role: normative
    status: modeled
    span:
      start: 0
      end: 335
origin:
  kind: agent
  recorded_at: '2026-11-08T00:35:50.612Z'
id: REQ-e2e-fixture-serves-bond-data
type: req
---
The end-to-end institution fixture must serve the treasury-bonds connector's sources in their real formats: the Ministry of Finance's gov.pl page and dane.gov.pl resource list leading to a synthetic .xls with each series' registry ISIN, rates and margin, withheld until the suite publishes it, GUS's CPI and NBP's interest rate archive.

## Context

D-200 criterion 5 is an end-to-end scenario: the data path's treasury-bonds instance reads the fixture instead of gov.pl, dane.gov.pl, GUS and NBP (testing.md: external institutions are fixture servers in the cluster; D-74 in-cluster services). Withholding the data first lets the suite book lots under provisional ISINs and see them re-keyed (D-214); the CPI is the demo dataset's so the run's readings agree with the demo's whichever source the core values bonds from (D-94). Modeled by Claude Code for D-200 criterion 5's end-to-end scenario; the author has not reviewed the wording.

## Source

> The end-to-end institution fixture must serve the treasury-bonds connector's sources in their real formats: the Ministry of Finance's gov.pl page and dane.gov.pl resource list leading to a synthetic .xls with each series' registry ISIN, rates and margin, withheld until the suite publishes it, GUS's CPI and NBP's interest rate archive.

docs/DESIGN.md#D-200
