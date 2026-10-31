---
title: The Ghostfolio plugin must remove the bond assets an earlier version of the plugin left after the import, like any stale asset.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-8
  - lane:strict
text_ref: design-decisions:D-64#8
origin:
  kind: import
  ref: design-decisions:D-64#8
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The Ghostfolio plugin must remove the bond assets an earlier version of the plugin left after the import, like any stale asset.
logic_claims:
  - CLAIM-24F2745EAA81E321
semantic_clauses:
  - The Ghostfolio plugin must remove the bond assets an earlier version of the plugin left after the import, like any stale asset.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 92adf5a53d8a667d888c38be1b041a04b7c0d7cbd76302b50f4fee6d512ea1cb
semantic_inventory:
  - claim_key: CLAIM-24F2745EAA81E321
    claim_text: The Ghostfolio plugin must remove the bond assets an earlier version of the plugin left after the import, like any stale asset
    role: normative
    status: modeled
    span:
      start: 0
      end: 126
id: REQ-views-ghostfolio-earlier-bond-assets-removed
type: req
---
The Ghostfolio plugin must remove the bond assets an earlier version of the plugin left after the import, like any stale asset.

## Context

An install that synced with an earlier version holds activities on bond assets that version created. Since each new bond asset's name carries a suffix no earlier version used (REQ-views-ghostfolio-bond-asset-fresh-name), no import lands on an old asset, so the old ones need no exception to import-first (REQ-views-ghostfolio-import-first-blocking-delete): their activities are stale and go after the import, and Ghostfolio deletes an asset along with its last activity. The implementation had deleted them before the import as a second exception (question #12 on docs/implementation-questions.md); the author amended D-64 on 2026-10-31 while answering the implementation's questions so that import-first keeps its one exception.

## Source

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
>   - A bond is a `MANUAL` asset whose symbol is a UUID derived from the plugin instance and its series code, so two installs that share one Ghostfolio never collide. Its name is the code plus a suffix no earlier version used, because Ghostfolio's import reuses an existing `MANUAL` asset of the same name; so an import never lands on an asset left by an earlier version, and those are removed after the import like any stale one. Every sync sends its full price history.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#8
