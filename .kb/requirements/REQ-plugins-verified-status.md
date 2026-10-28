---
title: The operator must record in each Plugin's status whether its image is verified or only signed by a trusted signer, and the UI must show that status.
status: open
text_ref: design-decisions:D-56#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-56-2
  - lane:strict
semantic_text: The operator must record in each Plugin's status whether its image is verified or only signed by a trusted signer, and the UI must show that status.
logic_claims:
  - CLAIM-BEF4182818AA2930
semantic_clauses:
  - The operator must record in each Plugin's status whether its image is verified or only signed by a trusted signer, and the UI must show that status.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 68239c10b531b45e9997fecf9f8da2df2676556e6197cb850b384e9e3ba5ac44
semantic_inventory:
  - claim_key: CLAIM-BEF4182818AA2930
    claim_text: The operator must record in each Plugin's status whether its image is verified or only signed by a trusted signer, and the UI must show that status
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-56#2
  recorded_at: '2026-10-28T06:09:27.982Z'
id: REQ-plugins-verified-status
type: req
---
The operator must record in each Plugin's status whether its image is verified or only signed by a trusted signer, and the UI must show that status.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> The operator verifies every plugin image's signature itself before deploying it, so a `Plugin` applied through GitOps or plain `kubectl apply` is checked too. It uses a bundled trust root, so verification works offline. It records in the `Plugin`'s status whether the image is verified (D-44) or only signed by a trusted signer, and the UI shows that status. The CLI's check (D-46) is a preview of the same check.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-56#2
