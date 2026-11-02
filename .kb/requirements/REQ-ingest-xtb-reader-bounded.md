---
title: The XTB workbook reader must stream every part and keep only cells with a value, and must refuse with ErrTooLarge a workbook with more than 1,000,000 such cells, 1,000,000 shared strings, 32 MiB of their text, 100 sheets or 200,000 rows a sheet, a run of more than 256 KiB between markup characters, or elements nested deeper than 32.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: The XTB workbook reader must stream every part and keep only cells with a value, and must refuse with ErrTooLarge a workbook with more than 1,000,000 such cells, 1,000,000 shared strings, 32 MiB of their text, 100 sheets or 200,000 rows a sheet, a run of more than 256 KiB between markup characters, or elements nested deeper than 32.
logic_claims:
  - CLAIM-2C0C6EFD837FC09D
semantic_clauses:
  - The XTB workbook reader must stream every part and keep only cells with a value, and must refuse with ErrTooLarge a workbook with more than 1,000,000 such cells, 1,000,000 shared strings, 32 MiB of their text, 100 sheets or 200,000 rows a sheet, a run of more than 256 KiB between markup characters, or elements nested deeper than 32.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3fafd5eec230f661c8b037de950dcef7a020c5cf3e5e9a04bdf0fb6acb1ff9e1
semantic_inventory:
  - claim_key: CLAIM-2C0C6EFD837FC09D
    claim_text: The XTB workbook reader must stream every part and keep only cells with a value, and must refuse with ErrTooLarge a workbook with more than 1,000,000 such cells, 1,000,000 shared strings, 32 MiB of their text, 100 sheets or 200,000 rows a sheet, a run of more than 256 KiB between markup characters, or elements nested deeper than 32
    role: normative
    status: modeled
    span:
      start: 0
      end: 333
origin:
  kind: agent
  recorded_at: '2026-11-02T20:57:39.664Z'
id: REQ-ingest-xtb-reader-bounded
type: req
---
The XTB workbook reader must stream every part and keep only cells with a value, and must refuse with ErrTooLarge a workbook with more than 1,000,000 such cells, 1,000,000 shared strings, 32 MiB of their text, 100 sheets or 200,000 rows a sheet, a run of more than 256 KiB between markup characters, or elements nested deeper than 32.

## Context

The reader padded every row to its last cell's column, so one <c r="XFD1"/> per row made 16,384 cells a row: a 0.4 KB zip of 1,000 such rows allocated 1.66 GB, against the core's 256 MiB memory limit, and the whole sheet and shared strings were unmarshalled at once. An export of 50,000 operations stays well within the caps. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p0-dos), from the security review's finding P0-3. The author made the fix a decision, D-184. No firm requirement for D-184 existed in kibi on this branch when it was built.

## Source

> The core bounds every request body and uploaded file before authenticating or parsing it, and every parser bounds what it allocates, so no unauthenticated request or small file can exhaust its memory.

docs/DESIGN.md#D-184

> GR-12: never log amounts, holdings, account numbers or tokens (docs/DESIGN.md#GR-12, docs/dev/code-organization.md#rules).
