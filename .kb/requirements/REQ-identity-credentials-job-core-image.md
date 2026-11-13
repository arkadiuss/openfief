---
title: A one-shot Job, run from the core's own signed image, generates the platform credentials.
status: open
text_ref: design-decisions:D-216#2
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-216-2
semantic_text: A one-shot Job, run from the core's own signed image, generates the platform credentials.
logic_claims:
  - CLAIM-03E2B307AEB2EB4B
semantic_clauses:
  - A one-shot Job, run from the core's own signed image, generates the platform credentials.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: aae819a9321c2dd1c8b7d7ead1f1b3269b9ec6fb8071551fe42d8bdd9c9a639f
semantic_inventory:
  - claim_key: CLAIM-03E2B307AEB2EB4B
    claim_text: A one-shot Job, run from the core's own signed image, generates the platform credentials
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 88
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-216#2
  recorded_at: '2026-11-13T07:42:49.836Z'
id: REQ-identity-credentials-job-core-image
type: req
---
A one-shot Job, run from the core's own signed image, generates the platform credentials.

## Context

Part of D-216. Our images are signed by images.yml and pinned by the release bundle to the digest that was signed (D-17), so a Job that runs the core's image adds no third-party image to the install. The decision states no further reason. The author recorded D-216 in DESIGN.md.

## Source

> The bootstrap install generates the platform credentials nothing else makes, such as SpiceDB's preshared key and its database role's password (D-164): a one-shot Job, run from the core's own signed image, creates each Secret only if it is missing and never reads, changes or rotates one, so an owner who feeds them from their own tooling first (D-29) keeps theirs. Its Role may only create Secrets in the `openfief` namespace.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-216#2