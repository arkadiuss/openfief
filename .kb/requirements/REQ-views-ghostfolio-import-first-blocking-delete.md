---
title: Each Ghostfolio sync must first delete the stale activities that Ghostfolio's duplicate check would match against an import, then import missing activities, and only then delete the other stale ones.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-1
  - lane:strict
text_ref: design-decisions:D-64#1
origin:
  kind: import
  ref: design-decisions:D-64#1
  recorded_at: '2026-10-29T10:21:58.896Z'
semantic_text: Each Ghostfolio sync must first delete the stale activities that Ghostfolio's duplicate check would match against an import, then import missing activities, and only then delete the other stale ones.
logic_claims:
  - CLAIM-96669E061B9B47A2
semantic_clauses:
  - Each Ghostfolio sync must first delete the stale activities that Ghostfolio's duplicate check would match against an import, then import missing activities, and only then delete the other stale ones.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 362adb3fe5f8b0e45f038059975542a40e2fd1e07591c4b18c45b9973a4436c9
semantic_inventory:
  - claim_key: CLAIM-96669E061B9B47A2
    claim_text: Each Ghostfolio sync must first delete the stale activities that Ghostfolio's duplicate check would match against an import, then import missing activities, and only then delete the other stale ones
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
id: REQ-views-ghostfolio-import-first-blocking-delete
type: req
---
Each Ghostfolio sync must first delete the stale activities that Ghostfolio's duplicate check would match against an import, then import missing activities, and only then delete the other stale ones.

## Why this changed

The author amended D-64: each Ghostfolio sync still imports missing activities before deleting stale ones, but a stale activity that Ghostfolio's duplicate check would match against an import is now deleted first, or the sync would never settle. REQ-views-ghostfolio-import-before-delete required every import to come before every delete, with no exception, so this requirement replaces it.

## Context

Ghostfolio deletes an asset and its price history along with its last activity, which is why D-64 imports before it deletes. The author amended D-64 to follow the Ghostfolio plugin as built, and recorded D-108 with it, after the design session that answered the implementation's questions.

## Source

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
>   - Each sync imports missing activities first, then deletes stale ones, because Ghostfolio deletes an asset and its price history along with its last activity. The one exception is a stale activity that Ghostfolio's duplicate check would match against an import: it is deleted first, or the sync would never settle. A bulk delete always carries an account filter; an unfiltered one deletes every activity the user has.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#1
