---
title: Ingest Xtb Reader
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.xtb_reader
canonical_key: ingest.xtb_reader
claim_key: CLAIM-2C0C6EFD837FC09D
claim_text: The XTB workbook reader must stream every part and keep only cells with a value, and must refuse with ErrTooLarge a workbook with more than 1,000,000 such cells, 1,000,000 shared strings, 32 MiB of their text, 100 sheets or 200,000 rows a sheet, a run of more than 256 KiB between markup characters, or elements nested deeper than 32.
origin:
  kind: agent
  recorded_at: '2026-11-02T20:57:35.294Z'
id: FACT-SUBJECT-429F29A98E5702B1
type: fact
---
