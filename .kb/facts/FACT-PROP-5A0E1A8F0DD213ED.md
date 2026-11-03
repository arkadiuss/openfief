---
title: Bounded By Pod Memory = true
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
property_key: bounded_by_pod_memory
canonical_key: mcp-kibi-model-requirement:ingest.xtb_reader:bounded_by_pod_memory:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FC25ECCA34F80EEC
claim_text: The XTB workbook reader must refuse a file over 16 MiB or with more than 1,000 zip entries, counted before the zip is opened; must read each sheet's part once, refusing a workbook that lists two sheets on one part or one sheet name twice; must keep only rows and cells with a value, at most 500,000 cells, 500,000 shared strings and 16 MiB of their text; and must refuse any XML token longer than 256 KiB as written.
origin:
  kind: agent
  recorded_at: '2026-11-03T07:18:02.025Z'
id: FACT-PROP-5A0E1A8F0DD213ED
type: fact
---
