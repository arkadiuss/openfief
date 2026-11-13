---
title: The Job never reads, changes or rotates a Secret, so an owner who feeds the platform credentials from their own tooling first (D-29) keeps theirs.
status: open
text_ref: design-decisions:D-216#4
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-216-4
semantic_text: The Job never reads, changes or rotates a Secret, so an owner who feeds the platform credentials from their own tooling first (D-29) keeps theirs.
logic_claims:
  - CLAIM-932B6DDB5213DA5E
semantic_clauses:
  - The Job never reads, changes or rotates a Secret, so an owner who feeds the platform credentials from their own tooling first (D-29) keeps theirs.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 76d10fb077aba9923c817cc0e277523ef89a5052de27c0f010dfc934349997d0
semantic_inventory:
  - claim_key: CLAIM-932B6DDB5213DA5E
    claim_text: The Job never reads, changes or rotates a Secret, so an owner who feeds the platform credentials from their own tooling first (D-29) keeps theirs
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 145
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-216#4
  recorded_at: '2026-11-13T07:42:57.210Z'
id: REQ-identity-credentials-never-read-change-rotate
type: req
---
The Job never reads, changes or rotates a Secret, so an owner who feeds the platform credentials from their own tooling first (D-29) keeps theirs.

## Context

Part of D-216, after D-29: credentials are Kubernetes Secrets fed by whatever the owner already runs. The Job only calls create, so running it again after an upgrade or a restore rotates nothing. The author recorded D-216 in DESIGN.md.

## Source

> The bootstrap install generates the platform credentials nothing else makes, such as SpiceDB's preshared key and its database role's password (D-164): a one-shot Job, run from the core's own signed image, creates each Secret only if it is missing and never reads, changes or rotates one, so an owner who feeds them from their own tooling first (D-29) keeps theirs. Its Role may only create Secrets in the `openfief` namespace.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-216#4