---
title: A Ghostfolio sync must delete an activity on a MANUAL asset an earlier version left after the import, like any stale activity, and delete before the import only the stale activities an import duplicates.
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: A Ghostfolio sync must delete an activity on a MANUAL asset an earlier version left after the import, like any stale activity, and delete before the import only the stale activities an import duplicates.
logic_claims:
  - CLAIM-5909FE07AD300656
semantic_clauses:
  - A Ghostfolio sync must delete an activity on a MANUAL asset an earlier version left after the import, like any stale activity, and delete before the import only the stale activities an import duplicates.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5510488d51237c1b8610fd6e6b1c4449dc4a6a5ddc1ee65eb437c93f3db49fd4
semantic_inventory:
  - claim_key: CLAIM-5909FE07AD300656
    claim_text: A Ghostfolio sync must delete an activity on a MANUAL asset an earlier version left after the import, like any stale activity, and delete before the import only the stale activities an import duplicates
    role: normative
    status: modeled
    span:
      start: 0
      end: 202
origin:
  kind: agent
  recorded_at: '2026-10-31T15:07:58.535Z'
id: REQ-views-ghostfolio-leftover-assets-after-import
type: req
---
A Ghostfolio sync must delete an activity on a MANUAL asset an earlier version left after the import, like any stale activity, and delete before the import only the stale activities an import duplicates.

## Context

D-64 amended: import-first keeps its single exception; the pre-import delete of leftover MANUAL assets the previous change added is gone (implementation question 12, now decided). Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> A bond is a `MANUAL` asset whose symbol is a UUID derived from the plugin instance and its series code, so two installs that share one Ghostfolio never collide. Its name is the code plus a suffix no earlier version used, because Ghostfolio's import reuses an existing `MANUAL` asset of the same name; so an import never lands on an asset left by an earlier version, and those are removed after the import like any stale one.

docs/DESIGN.md#D-64
