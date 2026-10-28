---
title: An index reading must name CPI or the NBP reference rate, a date and a source, and a CPI reading must be dated on the first of the month it measures; when any reading in a delivery is invalid, the core must fail the batch and import none of it.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: An index reading must name CPI or the NBP reference rate, a date and a source, and a CPI reading must be dated on the first of the month it measures; when any reading in a delivery is invalid, the core must fail the batch and import none of it.
logic_claims:
  - CLAIM-B131C15AFB7638FD
semantic_clauses:
  - An index reading must name CPI or the NBP reference rate, a date and a source, and a CPI reading must be dated on the first of the month it measures; when any reading in a delivery is invalid, the core must fail the batch and import none of it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f31308c160ac4d7f417eb7cf347aa24a25f44cb7c69c78c788414376b35ae5f0
semantic_inventory:
  - claim_key: CLAIM-B131C15AFB7638FD
    claim_text: An index reading must name CPI or the NBP reference rate, a date and a source, and a CPI reading must be dated on the first of the month it measures; when any reading in a delivery is invalid, the core must fail the batch and import none of it
    role: normative
    status: modeled
    span:
      start: 0
      end: 243
origin:
  kind: agent
  recorded_at: '2026-10-28T08:56:32.857Z'
id: REQ-ingest-index-reading-valid
type: req
---
An index reading must name CPI or the NBP reference rate, a date and a source, and a CPI reading must be dated on the first of the month it measures; when any reading in a delivery is invalid, the core must fail the batch and import none of it.

## Context

D-36 derives valuations from index readings among other inputs, and jurisdiction/pl takes CPI year on year and the NBP reference rate as yearly fractions (pl.Indices). A reading is keyed by index and date, so a CPI reading must name its month unambiguously. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time.

docs/DESIGN.md D-36
