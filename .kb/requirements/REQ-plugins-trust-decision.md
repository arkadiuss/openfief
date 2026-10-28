---
title: The trust policy must call an image verified only when a catalog entry pins its digest and that entry's publisher signed it, unverified when any other trusted signer signed it, and must refuse it otherwise.
status: open
text_ref: docs/DESIGN.md#D-44
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-44
  - lane:strict
semantic_text: The trust policy must call an image verified only when a catalog entry pins its digest and that entry's publisher signed it, unverified when any other trusted signer signed it, and must refuse it otherwise.
logic_claims:
  - CLAIM-8FB52869CF7C6420
semantic_clauses:
  - The trust policy must call an image verified only when a catalog entry pins its digest and that entry's publisher signed it, unverified when any other trusted signer signed it, and must refuse it otherwise.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 66fd419468da9bb778f5059126c8aa3b4c4fbf4cabf1a025353342c3573f17d5
semantic_inventory:
  - claim_key: CLAIM-8FB52869CF7C6420
    claim_text: The trust policy must call an image verified only when a catalog entry pins its digest and that entry's publisher signed it, unverified when any other trusted signer signed it, and must refuse it otherwise
    role: normative
    status: modeled
    span:
      start: 0
      end: 205
origin:
  kind: agent
  recorded_at: '2026-10-28T08:51:47.543Z'
id: REQ-plugins-trust-decision
type: req
---
The trust policy must call an image verified only when a catalog entry pins its digest and that entry's publisher signed it, unverified when any other trusted signer signed it, and must refuse it otherwise.

## Context

D-44 makes verified mean a catalog entry pins the publisher's signed digest and signer identity; D-45 trusts CI, catalog publishers and owner signers, whose plugins show as unverified; D-57 refuses everything else with no override. The catalog pins the digest only, so a mirror of the same content is verified too. Refines REQ-plugins-verified-entry-pins, REQ-plugins-trusted-signers-only and REQ-plugins-unsigned-never-deployed. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> A verified plugin is one whose version the author reviewed: its catalog entry pins the publisher's signed image digest and the publisher's signer identity.

docs/DESIGN.md D-44

> An image without a valid signature from a trusted signer is never deployed, and there is no override.

docs/DESIGN.md D-57
