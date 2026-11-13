---
title: The Job creates each Secret only if it is missing.
status: open
text_ref: design-decisions:D-216#3
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-216-3
semantic_text: The Job creates each Secret only if it is missing.
logic_claims:
  - CLAIM-50AAE59870B78646
semantic_clauses:
  - The Job creates each Secret only if it is missing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ca6ed0218447c4e04c7c8842d4b53d4958e048ce7d663a65b3a6e8af80de4334
semantic_inventory:
  - claim_key: CLAIM-50AAE59870B78646
    claim_text: The Job creates each Secret only if it is missing
    role: normative
    status: modeled
    span:
      start: 0
      end: 49
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-216#3
  recorded_at: '2026-11-13T07:42:53.958Z'
id: REQ-identity-credentials-create-only-if-missing
type: req
---
The Job creates each Secret only if it is missing.

## Context

Part of D-216: the Job creates each Secret through the API server and a Secret that exists answers 409 and is kept. The decision states no further reason. The author recorded D-216 in DESIGN.md.

## Source

> The bootstrap install generates the platform credentials nothing else makes, such as SpiceDB's preshared key and its database role's password (D-164): a one-shot Job, run from the core's own signed image, creates each Secret only if it is missing and never reads, changes or rotates one, so an owner who feeds them from their own tooling first (D-29) keeps theirs. Its Role may only create Secrets in the `openfief` namespace.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-216#3