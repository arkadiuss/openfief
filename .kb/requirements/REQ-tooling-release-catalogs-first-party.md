---
title: The release workflow must add each first-party plugin to the catalog with the image digest it signed.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-88-1
  - lane:strict
text_ref: design-decisions:D-88#1
origin:
  kind: import
  ref: design-decisions:D-88#1
  recorded_at: '2026-10-29T08:30:24.091Z'
semantic_text: The release workflow must add each first-party plugin to the catalog with the image digest it signed.
logic_claims:
  - CLAIM-C481A72FBB4FF461
semantic_clauses:
  - The release workflow must add each first-party plugin to the catalog with the image digest it signed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 66e3260c09fb541829b7123177d072448de1e592c179d2247fe8284c8f0e3703
semantic_inventory:
  - claim_key: CLAIM-C481A72FBB4FF461
    claim_text: The release workflow must add each first-party plugin to the catalog with the image digest it signed
    role: normative
    status: modeled
    span:
      start: 0
      end: 100
id: REQ-tooling-release-catalogs-first-party
type: req
---
The release workflow must add each first-party plugin to the catalog with the image digest it signed.

## Context

D-43 puts first-party and verified plugins in one catalog file released with the product (REQ-plugins-catalog-in-repo), and D-44 has a verified entry pin the publisher's signed image digest and signer (REQ-plugins-verified-entry-pins). D-88 has the release workflow write those entries for first-party plugins, with the digest it signed, so the catalog and the released images agree. REQ-tooling-release-started-by-author and REQ-plugins-first-party-verified record the rest of D-88. The author recorded D-88 to D-90 after the design session that answered the implementation's questions.

## Source

> The release workflow adds each first-party plugin to the catalog with the image digest it signed. The author starts every release by hand, and that release is the review D-44 requires, so first-party plugins show as verified.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-88#1
