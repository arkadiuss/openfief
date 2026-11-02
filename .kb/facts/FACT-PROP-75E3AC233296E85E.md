---
title: Allocation Bounded = true
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.xtb_reader
property_key: allocation_bounded
canonical_key: mcp-kibi-model-requirement:ingest.xtb_reader:allocation_bounded:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2C0C6EFD837FC09D
claim_text: The XTB workbook reader must stream every part and keep only cells with a value, and must refuse with ErrTooLarge a workbook with more than 1,000,000 such cells, 1,000,000 shared strings, 32 MiB of their text, 100 sheets or 200,000 rows a sheet, a run of more than 256 KiB between markup characters, or elements nested deeper than 32.
origin:
  kind: agent
  recorded_at: '2026-11-02T20:57:37.272Z'
id: FACT-PROP-75E3AC233296E85E
type: fact
---
