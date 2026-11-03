---
title: The XTB workbook reader must refuse a file over 16 MiB or with more than 1,000 zip entries, counted before the zip is opened; must read each sheet's part once, refusing a workbook that lists two sheets on one part or one sheet name twice; must keep only rows and cells with a value, at most 500,000 cells, 500,000 shared strings and 16 MiB of their text; and must refuse any XML token longer than 256 KiB as written.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: The XTB workbook reader must refuse a file over 16 MiB or with more than 1,000 zip entries, counted before the zip is opened; must read each sheet's part once, refusing a workbook that lists two sheets on one part or one sheet name twice; must keep only rows and cells with a value, at most 500,000 cells, 500,000 shared strings and 16 MiB of their text; and must refuse any XML token longer than 256 KiB as written.
logic_claims:
  - CLAIM-FC25ECCA34F80EEC
semantic_clauses:
  - The XTB workbook reader must refuse a file over 16 MiB or with more than 1,000 zip entries, counted before the zip is opened; must read each sheet's part once, refusing a workbook that lists two sheets on one part or one sheet name twice; must keep only rows and cells with a value, at most 500,000 cells, 500,000 shared strings and 16 MiB of their text; and must refuse any XML token longer than 256 KiB as written.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 87b95e5dbb3f1628d88421f24602b0f93899bd5107613cb370aff26c32cd533a
semantic_inventory:
  - claim_key: CLAIM-FC25ECCA34F80EEC
    claim_text: The XTB workbook reader must refuse a file over 16 MiB or with more than 1,000 zip entries, counted before the zip is opened; must read each sheet's part once, refusing a workbook that lists two sheets on one part or one sheet name twice; must keep only rows and cells with a value, at most 500,000 cells, 500,000 shared strings and 16 MiB of their text; and must refuse any XML token longer than 256 KiB as written
    role: normative
    status: modeled
    span:
      start: 0
      end: 415
origin:
  kind: agent
  recorded_at: '2026-11-03T07:18:06.364Z'
id: REQ-ingest-xtb-reader-bounded-by-pod-memory
type: req
---
The XTB workbook reader must refuse a file over 16 MiB or with more than 1,000 zip entries, counted before the zip is opened; must read each sheet's part once, refusing a workbook that lists two sheets on one part or one sheet name twice; must keep only rows and cells with a value, at most 500,000 cells, 500,000 shared strings and 16 MiB of their text; and must refuse any XML token longer than 256 KiB as written.

## Context

Recorded with default firmness by the implementation agent of the second security review's fixes (fix/review2-r1-r4), from finding R2. The old run-of-bytes heuristic reset on '>', which is legal inside an attribute value, so a 331 KiB file with a 120 MiB attribute grew the heap by 372 MiB; 100 sheets pointing at one part of 200,000 empty rows read that part 100 times and kept every empty row, 1.5 GiB; and zip.NewReader allocated for every entry before the entry cap was checked. The caps are derived from the xtb plugin pod's 256 MiB memory limit, which the operator sets for every plugin container; the namespace LimitRange's 1 GiB default applies only to containers without a limit of their own. Measured now: every hostile case grows the heap by under 16 MiB, and the largest workbook the reader keeps by about 70 MiB.

## Why this changed

It supersedes REQ-ingest-xtb-reader-bounded: that requirement bounded runs of bytes between markup characters, which an attribute value full of '>' defeats, let many sheets read one part, kept empty rows, and opened the zip before counting its entries; its cell, string and text caps let a small file reach about 140 MiB of heap, too close to the plugin pod's 256 MiB limit.

## Source

> The core bounds every request body and uploaded file before authenticating or parsing it, and every parser bounds what it allocates, so no unauthenticated request or small file can exhaust its memory.

docs/DESIGN.md#D-184
