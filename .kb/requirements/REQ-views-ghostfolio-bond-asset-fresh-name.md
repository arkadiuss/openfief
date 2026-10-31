---
title: The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code and whose name is the series code plus a suffix no earlier version of the plugin used, so that an import never lands on an asset an earlier version left.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-6
  - lane:strict
text_ref: design-decisions:D-64#6
origin:
  kind: import
  ref: design-decisions:D-64#6
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code and whose name is the series code plus a suffix no earlier version of the plugin used, so that an import never lands on an asset an earlier version left.
logic_claims:
  - CLAIM-C91D7DC152542B04
semantic_clauses:
  - The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code and whose name is the series code plus a suffix no earlier version of the plugin used, so that an import never lands on an asset an earlier version left.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 3bfb3817d70eeb8f4cdcded446fe3ba1806fe7d163fc7dd618b161e26d136ff6
semantic_inventory:
  - claim_key: CLAIM-C91D7DC152542B04
    claim_text: The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code and whose name is the series code plus a suffix no earlier version of the plugin used, so that an import never lands on an asset an earlier version left
    role: normative
    status: modeled
    span:
      start: 0
      end: 285
id: REQ-views-ghostfolio-bond-asset-fresh-name
type: req
---
The Ghostfolio plugin must make each bond a MANUAL asset whose symbol is a UUID derived from the plugin instance and its series code and whose name is the series code plus a suffix no earlier version of the plugin used, so that an import never lands on an asset an earlier version left.

## Why this changed

REQ-views-ghostfolio-bond-asset-symbol named each bond's MANUAL asset by its series code alone. Ghostfolio's import reuses an existing MANUAL asset of the same name, so once D-64 derived the symbol from the plugin instance as well, every import of a new asset landed on the asset an earlier version had left under the same code, its dry run was refused and the sync never settled. The implementation worked around it with a second exception to import-first (question #12 on docs/implementation-questions.md); on 2026-10-31 the author amended D-64 instead, so the name carries a suffix no earlier version used and import-first keeps its one exception.

## Context

D-64 makes each bond a MANUAL asset whose UUID symbol comes from the plugin instance and its series code, so two installs that share one Ghostfolio never collide. The assets an earlier version left are removed after the import like any stale one (REQ-views-ghostfolio-earlier-bond-assets-removed), and each sync still sends a bond's full price history. The author decided this on 2026-10-31 while answering the implementation's questions.

## Source

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
>   - A bond is a `MANUAL` asset whose symbol is a UUID derived from the plugin instance and its series code, so two installs that share one Ghostfolio never collide. Its name is the code plus a suffix no earlier version used, because Ghostfolio's import reuses an existing `MANUAL` asset of the same name; so an import never lands on an asset left by an earlier version, and those are removed after the import like any stale one. Every sync sends its full price history.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#6
