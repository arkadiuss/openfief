---
title: The treasury bond connector must read each series' registry ISIN from the Kod ISIN column of the ministry's workbook, a valid Polish ISIN that is no provisional one, a series listed before its rates included, and refuse a workbook listing one ISIN for two series.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: The treasury bond connector must read each series' registry ISIN from the Kod ISIN column of the ministry's workbook, a valid Polish ISIN that is no provisional one, a series listed before its rates included, and refuse a workbook listing one ISIN for two series.
logic_claims:
  - CLAIM-2814515260C23B95
semantic_clauses:
  - The treasury bond connector must read each series' registry ISIN from the Kod ISIN column of the ministry's workbook, a valid Polish ISIN that is no provisional one, a series listed before its rates included, and refuse a workbook listing one ISIN for two series.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a4f2b518ef04c0f34767369f598a9a3f126fcd1441136a77e73ae71dbd08ed05
semantic_inventory:
  - claim_key: CLAIM-2814515260C23B95
    claim_text: The treasury bond connector must read each series' registry ISIN from the Kod ISIN column of the ministry's workbook, a valid Polish ISIN that is no provisional one, a series listed before its rates included, and refuse a workbook listing one ISIN for two series
    role: normative
    status: modeled
    span:
      start: 0
      end: 262
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:04.413Z'
id: REQ-treasurybonds-reads-registry-isin
type: req
---
The treasury bond connector must read each series' registry ISIN from the Kod ISIN column of the ministry's workbook, a valid Polish ISIN that is no provisional one, a series listed before its rates included, and refuse a workbook listing one ISIN for two series.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97), never one derived from its series code; the series code is its ticker.

docs/DESIGN.md#D-214
