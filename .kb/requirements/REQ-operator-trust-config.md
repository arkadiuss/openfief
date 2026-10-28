---
title: The operator must build its trust policy from OpenFief's CI identity, the catalog embedded in its binary and an optional owner signer file, must refuse to start when that file is invalid, and must offer no setting that deploys an image the policy refuses.
status: open
text_ref: docs/DESIGN.md#D-45
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-45
  - lane:strict
semantic_text: The operator must build its trust policy from OpenFief's CI identity, the catalog embedded in its binary and an optional owner signer file, must refuse to start when that file is invalid, and must offer no setting that deploys an image the policy refuses.
logic_claims:
  - CLAIM-9F39D68871C76E2A
semantic_clauses:
  - The operator must build its trust policy from OpenFief's CI identity, the catalog embedded in its binary and an optional owner signer file, must refuse to start when that file is invalid, and must offer no setting that deploys an image the policy refuses.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1511e5cc256fa7ad4cb0c980851b0b9a495150aed12a75aaf4202c1409d0caee
semantic_inventory:
  - claim_key: CLAIM-9F39D68871C76E2A
    claim_text: The operator must build its trust policy from OpenFief's CI identity, the catalog embedded in its binary and an optional owner signer file, must refuse to start when that file is invalid, and must offer no setting that deploys an image the policy refuses
    role: normative
    status: modeled
    span:
      start: 0
      end: 254
origin:
  kind: agent
  recorded_at: '2026-10-28T09:29:44.785Z'
id: REQ-operator-trust-config
type: req
---
The operator must build its trust policy from OpenFief's CI identity, the catalog embedded in its binary and an optional owner signer file, must refuse to start when that file is invalid, and must offer no setting that deploys an image the policy refuses.

## Context

D-45 lists exactly three kinds of trusted signer; D-43 releases the catalog with the product; D-57 rules out any override. Owner signers come from the openfief-trusted-signers ConfigMap, mounted as a file and read at start. An invalid file stops the operator rather than being skipped, so a typo can neither widen nor silently narrow trust. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> The operator trusts three kinds of signer: OpenFief's CI, the publishers pinned in the catalog, and signers the owner adds.

docs/DESIGN.md D-45

> An image without a valid signature from a trusted signer is never deployed, and there is no override.

docs/DESIGN.md D-57
