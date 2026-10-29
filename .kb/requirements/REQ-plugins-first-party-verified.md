---
title: First-party plugins must show as verified, with the author's hand-started release counting as the review D-44 requires.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-88-3
  - lane:strict
text_ref: design-decisions:D-88#3
origin:
  kind: import
  ref: design-decisions:D-88#3
  recorded_at: '2026-10-29T08:31:35.947Z'
semantic_text: First-party plugins must show as verified, with the author's hand-started release counting as the review D-44 requires.
logic_claims:
  - CLAIM-518A0CDDBC665487
semantic_clauses:
  - First-party plugins must show as verified, with the author's hand-started release counting as the review D-44 requires.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e1a38d0dc5d64044a8d600ac70ba090205dc11ac4e5be60df2a30e734e2f688e
semantic_inventory:
  - claim_key: CLAIM-518A0CDDBC665487
    claim_text: First-party plugins must show as verified, with the author's hand-started release counting as the review D-44 requires
    role: normative
    status: modeled
    span:
      start: 0
      end: 118
id: REQ-plugins-first-party-verified
type: req
---
First-party plugins must show as verified, with the author's hand-started release counting as the review D-44 requires.

## Context

D-44 counts a plugin version as verified only when the author reviewed it and a catalog entry pins its signed digest (REQ-plugins-verified-entry-pins). D-88 settles how first-party plugins meet that: the author starts every release by hand (REQ-tooling-release-started-by-author), that release is the review, and the release workflow adds each first-party plugin to the catalog with the digest it signed (REQ-tooling-release-catalogs-first-party), so they show as verified. The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> The release workflow adds each first-party plugin to the catalog with the image digest it signed. The author starts every release by hand, and that release is the review D-44 requires, so first-party plugins show as verified.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-88#3
