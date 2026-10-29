---
title: Every OpenFief release must be started by hand by the author.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-88-2
  - lane:strict
text_ref: design-decisions:D-88#2
origin:
  kind: import
  ref: design-decisions:D-88#2
  recorded_at: '2026-10-29T08:30:57.692Z'
semantic_text: Every OpenFief release must be started by hand by the author.
logic_claims:
  - CLAIM-CCA3AF47B45939FC
semantic_clauses:
  - Every OpenFief release must be started by hand by the author.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 1a0a344363974cc727dfc713b6904044cc6a25860536246b04c3a16e46d70e4a
semantic_inventory:
  - claim_key: CLAIM-CCA3AF47B45939FC
    claim_text: Every OpenFief release must be started by hand by the author
    role: normative
    status: modeled
    span:
      start: 0
      end: 60
id: REQ-tooling-release-started-by-author
type: req
---
Every OpenFief release must be started by hand by the author.

## Context

D-44 counts a plugin version as verified only when the author reviewed it. D-88 makes the release itself that review for first-party plugins, which holds only because the author starts every release by hand rather than letting it run on a schedule or on every merge. REQ-tooling-release-catalogs-first-party and REQ-plugins-first-party-verified record the rest of D-88. The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> The release workflow adds each first-party plugin to the catalog with the image digest it signed. The author starts every release by hand, and that release is the review D-44 requires, so first-party plugins show as verified.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-88#2
