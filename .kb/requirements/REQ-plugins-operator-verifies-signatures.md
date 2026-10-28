---
title: The operator must verify every plugin image's signature before deploying it, including a Plugin applied without the CLI, and verification must work offline.
status: open
text_ref: design-decisions:D-56#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-56-1
  - lane:strict
semantic_text: The operator must verify every plugin image's signature before deploying it, including a Plugin applied without the CLI, and verification must work offline.
logic_claims:
  - CLAIM-FD6E0F61B0D0C94C
semantic_clauses:
  - The operator must verify every plugin image's signature before deploying it, including a Plugin applied without the CLI, and verification must work offline.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f689dc653c242c443bb62b5dd73abb7b8553d338b9b9c0facef37cd63400d980
semantic_inventory:
  - claim_key: CLAIM-FD6E0F61B0D0C94C
    claim_text: The operator must verify every plugin image's signature before deploying it, including a Plugin applied without the CLI, and verification must work offline
    role: normative
    status: modeled
    span:
      start: 0
      end: 155
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-56#1
  recorded_at: '2026-10-28T06:09:23.792Z'
id: REQ-plugins-operator-verifies-signatures
type: req
---
The operator must verify every plugin image's signature before deploying it, including a Plugin applied without the CLI, and verification must work offline.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> The operator verifies every plugin image's signature itself before deploying it, so a `Plugin` applied through GitOps or plain `kubectl apply` is checked too. It uses a bundled trust root, so verification works offline. It records in the `Plugin`'s status whether the image is verified (D-44) or only signed by a trusted signer, and the UI shows that status. The CLI's check (D-46) is a preview of the same check.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-56#1
