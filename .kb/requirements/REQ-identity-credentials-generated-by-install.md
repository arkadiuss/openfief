---
title: The bootstrap install generates the platform credentials nothing else makes.
status: open
text_ref: design-decisions:D-216#1
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-216-1
semantic_text: The bootstrap install generates the platform credentials nothing else makes.
logic_claims:
  - CLAIM-D5AB5221E57ABBEF
semantic_clauses:
  - The bootstrap install generates the platform credentials nothing else makes.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fbf1be1690fc04b355478879c473e704d24d487af1e24a97fb69ec87108abaf0
semantic_inventory:
  - claim_key: CLAIM-D5AB5221E57ABBEF
    claim_text: The bootstrap install generates the platform credentials nothing else makes
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 75
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-216#1
  recorded_at: '2026-11-13T07:42:45.934Z'
id: REQ-identity-credentials-generated-by-install
type: req
---
The bootstrap install generates the platform credentials nothing else makes.

## Context

Part of D-216, whose examples are SpiceDB's preshared key and its database role's password (D-164): CloudNativePG generates a password only for the owner of the database it creates, so SpiceDB's role password and the preshared key SpiceDB shares with the core (D-164) need a maker. The credentials Job of deploy/spicedb runs `core init-credentials` and makes both, once. The decision states no further reason. The author recorded D-216 in DESIGN.md.

## Source

> The bootstrap install generates the platform credentials nothing else makes, such as SpiceDB's preshared key and its database role's password (D-164): a one-shot Job, run from the core's own signed image, creates each Secret only if it is missing and never reads, changes or rotates one, so an owner who feeds them from their own tooling first (D-29) keeps theirs. Its Role may only create Secrets in the `openfief` namespace.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-216#1